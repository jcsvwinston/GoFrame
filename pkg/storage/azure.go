package storage

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"io"
	"net/url"
	"time"

	"github.com/Azure/azure-sdk-for-go/sdk/storage/azblob"
	"github.com/Azure/azure-sdk-for-go/sdk/storage/azblob/blob"
	"github.com/Azure/azure-sdk-for-go/sdk/storage/azblob/bloberror"
	"github.com/Azure/azure-sdk-for-go/sdk/storage/azblob/container"
	"github.com/Azure/azure-sdk-for-go/sdk/storage/azblob/sas"
)

// azureStore implements the Store interface for Azure Blob Storage.
type azureStore struct {
	client          *azblob.Client
	sharedKeyCred   *azblob.SharedKeyCredential
	containerName   string
	publicContainer string
	accountName     string
}

// Compile-time check that azureStore implements Store.
var _ Store = (*azureStore)(nil)

// NewAzureStore creates an Azure Blob client using shared key authentication.
func NewAzureStore(cfg AzureConfig) (Store, error) {
	accountName, err := cfg.AccountName.Resolve()
	if err != nil {
		return nil, fmt.Errorf("storage: resolve Azure account name: %w", err)
	}

	accountKey, err := cfg.AccountKey.Resolve()
	if err != nil {
		return nil, fmt.Errorf("storage: resolve Azure account key: %w", err)
	}

	if accountName == "" {
		return nil, errors.New("storage: Azure account name is empty after resolution")
	}
	if accountKey == "" {
		return nil, errors.New("storage: Azure account key is empty after resolution")
	}

	credential, err := azblob.NewSharedKeyCredential(accountName, accountKey)
	if err != nil {
		return nil, fmt.Errorf("storage: create Azure shared key credential: %w", err)
	}

	serviceURL := fmt.Sprintf("https://%s.blob.core.windows.net/", accountName)
	client, err := azblob.NewClientWithSharedKeyCredential(serviceURL, credential, nil)
	if err != nil {
		return nil, fmt.Errorf("storage: create Azure service client: %w", err)
	}

	return &azureStore{
		client:          client,
		sharedKeyCred:   credential,
		containerName:   cfg.Container,
		publicContainer: cfg.PublicContainer,
		accountName:     accountName,
	}, nil
}

// isNotFoundError checks if the error indicates a missing blob.
func isNotFoundError(err error) bool {
	if err == nil {
		return false
	}
	return bloberror.HasCode(err, bloberror.BlobNotFound)
}

// normalizeError converts Azure errors to storage package errors.
func normalizeError(key string, err error) error {
	if err == nil {
		return nil
	}
	if isNotFoundError(err) {
		return ErrNotFound(key)
	}
	return fmt.Errorf("storage: azure: %w", err)
}

// targetContainer determines which container to use based on visibility.
func (s *azureStore) targetContainer(opts PutOptions) string {
	if opts.Visibility == Public && s.publicContainer != "" {
		return s.publicContainer
	}
	return s.containerName
}

// Put uploads a file from an io.Reader to Azure Blob Storage.
func (s *azureStore) Put(ctx context.Context, key string, reader io.Reader, opts PutOptions) (ObjectInfo, error) {
	containerName := s.targetContainer(opts)

	data, err := io.ReadAll(reader)
	if err != nil {
		return ObjectInfo{}, fmt.Errorf("storage: read reader: %w", err)
	}

	uploadOpts := &azblob.UploadStreamOptions{}
	if opts.ContentType != "" {
		uploadOpts.HTTPHeaders = &blob.HTTPHeaders{
			BlobContentType: &opts.ContentType,
		}
	}

	_, err = s.client.UploadStream(ctx, containerName, key, bytes.NewReader(data), uploadOpts)
	if err != nil {
		return ObjectInfo{}, normalizeError(key, err)
	}

	propResp, err := s.client.DownloadStream(ctx, containerName, key, nil)
	if err != nil {
		return ObjectInfo{}, normalizeError(key, err)
	}
	defer propResp.Body.Close()

	info := ObjectInfo{
		Key:        key,
		Size:       *propResp.ContentLength,
		Visibility: opts.Visibility,
		Metadata:   opts.Metadata,
	}

	if propResp.ContentType != nil {
		info.ContentType = *propResp.ContentType
	}
	if propResp.LastModified != nil {
		info.UpdatedAt = *propResp.LastModified
	}

	return info, nil
}

// Get retrieves a file by key from Azure Blob Storage.
func (s *azureStore) Get(ctx context.Context, key string) (io.ReadCloser, ObjectInfo, error) {
	downloadResp, err := s.client.DownloadStream(ctx, s.containerName, key, nil)
	if err != nil {
		return nil, ObjectInfo{}, normalizeError(key, err)
	}

	info := ObjectInfo{
		Key:        key,
		Visibility: Private,
	}

	if downloadResp.ContentLength != nil {
		info.Size = *downloadResp.ContentLength
	}
	if downloadResp.ContentType != nil {
		info.ContentType = *downloadResp.ContentType
	}
	if downloadResp.LastModified != nil {
		info.UpdatedAt = *downloadResp.LastModified
	}

	return downloadResp.Body, info, nil
}

// Delete removes an object by key. Idempotent: no error if key doesn't exist.
func (s *azureStore) Delete(ctx context.Context, key string) error {
	_, err := s.client.DeleteBlob(ctx, s.containerName, key, nil)
	if err != nil {
		if isNotFoundError(err) {
			return nil
		}
		return normalizeError(key, err)
	}
	return nil
}

// Exists checks if a key exists in Azure Blob Storage.
func (s *azureStore) Exists(ctx context.Context, key string) (bool, error) {
	_, err := s.client.DownloadStream(ctx, s.containerName, key, &azblob.DownloadStreamOptions{
		Range: blob.HTTPRange{Count: 0},
	})
	if err != nil {
		if isNotFoundError(err) {
			return false, nil
		}
		return false, normalizeError(key, err)
	}
	return true, nil
}

// List returns objects with the given prefix and delimiter support.
func (s *azureStore) List(ctx context.Context, opts ListOptions) (ListResult, error) {
	var result ListResult

	if opts.Delimiter != "" {
		// Use hierarchy mode for directory-like listing.
		containerClient := s.client.ServiceClient().NewContainerClient(s.containerName)
		pager := containerClient.NewListBlobsHierarchyPager(opts.Delimiter, &container.ListBlobsHierarchyOptions{
			Prefix: &opts.Prefix,
			Marker: strPtr(opts.Marker),
		})

		if pager.More() {
			page, err := pager.NextPage(ctx)
			if err != nil {
				return result, fmt.Errorf("storage: azure list: %w", err)
			}

			for _, blobInfo := range page.Segment.BlobItems {
				item := ObjectInfo{Size: int64(0)}
				if blobInfo.Name != nil {
					item.Key = *blobInfo.Name
				}
				if blobInfo.Properties.ContentLength != nil {
					item.Size = *blobInfo.Properties.ContentLength
				}
				if blobInfo.Properties.ContentType != nil {
					item.ContentType = *blobInfo.Properties.ContentType
				}
				if blobInfo.Properties.LastModified != nil {
					item.UpdatedAt = *blobInfo.Properties.LastModified
				}
				result.Objects = append(result.Objects, item)
			}

			for _, prefixInfo := range page.Segment.BlobPrefixes {
				if prefixInfo.Name != nil {
					result.CommonPrefixes = append(result.CommonPrefixes, *prefixInfo.Name)
				}
			}

			if page.NextMarker != nil && *page.NextMarker != "" {
				result.NextMarker = *page.NextMarker
				result.Truncated = true
			}
		}
	} else {
		// Use flat listing.
		pager := s.client.NewListBlobsFlatPager(s.containerName, &azblob.ListBlobsFlatOptions{
			Prefix: &opts.Prefix,
			Marker: strPtr(opts.Marker),
		})

		if pager.More() {
			page, err := pager.NextPage(ctx)
			if err != nil {
				return result, fmt.Errorf("storage: azure list: %w", err)
			}

			for _, blobInfo := range page.Segment.BlobItems {
				item := ObjectInfo{Size: int64(0)}
				if blobInfo.Name != nil {
					item.Key = *blobInfo.Name
				}
				if blobInfo.Properties.ContentLength != nil {
					item.Size = *blobInfo.Properties.ContentLength
				}
				if blobInfo.Properties.ContentType != nil {
					item.ContentType = *blobInfo.Properties.ContentType
				}
				if blobInfo.Properties.LastModified != nil {
					item.UpdatedAt = *blobInfo.Properties.LastModified
				}
				result.Objects = append(result.Objects, item)
			}

			if page.NextMarker != nil && *page.NextMarker != "" {
				result.NextMarker = *page.NextMarker
				result.Truncated = true
			}
		}
	}

	return result, nil
}

func strPtr(s string) *string {
	if s == "" {
		return nil
	}
	return &s
}

// PublicURL returns a publicly accessible URL for a key.
func (s *azureStore) PublicURL(ctx context.Context, key string, opts URLConfig) (string, error) {
	if s.publicContainer == "" {
		return "", nil
	}

	// Check if the blob exists in the public container.
	_, err := s.client.DownloadStream(ctx, s.publicContainer, key, &azblob.DownloadStreamOptions{
		Range: blob.HTTPRange{Count: 0},
	})
	if err != nil {
		if isNotFoundError(err) {
			return "", nil
		}
		return "", normalizeError(key, err)
	}

	escapedKey := url.PathEscape(key)
	return fmt.Sprintf("https://%s.blob.core.windows.net/%s/%s", s.accountName, s.publicContainer, escapedKey), nil
}

// SignedURL returns a time-limited URL for accessing a private object.
func (s *azureStore) SignedURL(ctx context.Context, key string, expires time.Duration, opts URLConfig) (string, error) {
	if expires <= 0 {
		expires = 24 * time.Hour
	}

	containerClient := s.client.ServiceClient().NewContainerClient(s.containerName)
	blobClient := containerClient.NewBlobClient(key)

	startTime := time.Now().UTC()
	expiryTime := startTime.Add(expires)

	permissions := sas.BlobPermissions{Read: true}

	sasURL, err := blobClient.GetSASURL(permissions, expiryTime, &blob.GetSASURLOptions{
		StartTime: &startTime,
	})
	if err != nil {
		return "", fmt.Errorf("storage: azure generate SAS URL: %w", err)
	}

	return sasURL, nil
}

// Copy copies an object from srcKey to dstKey within the same container.
func (s *azureStore) Copy(ctx context.Context, srcKey, dstKey string) (ObjectInfo, error) {
	srcContainerClient := s.client.ServiceClient().NewContainerClient(s.containerName)
	srcBlobClient := srcContainerClient.NewBlobClient(srcKey)
	srcURL := srcBlobClient.URL()

	dstContainerClient := s.client.ServiceClient().NewContainerClient(s.containerName)
	dstBlobClient := dstContainerClient.NewBlobClient(dstKey)

	_, err := dstBlobClient.CopyFromURL(ctx, srcURL, nil)
	if err != nil {
		return ObjectInfo{}, normalizeError(srcKey, err)
	}

	// Get properties of the destination blob for info.
	propResp, err := dstBlobClient.GetProperties(ctx, nil)
	if err != nil {
		return ObjectInfo{}, normalizeError(dstKey, err)
	}

	info := ObjectInfo{
		Key:        dstKey,
		Visibility: Private,
	}

	if propResp.ContentLength != nil {
		info.Size = *propResp.ContentLength
	}
	if propResp.LastModified != nil {
		info.UpdatedAt = *propResp.LastModified
	}
	if propResp.ContentType != nil {
		info.ContentType = *propResp.ContentType
	}

	return info, nil
}

// Close releases resources held by the Azure store.
func (s *azureStore) Close() error {
	s.client = nil
	s.sharedKeyCred = nil
	return nil
}
