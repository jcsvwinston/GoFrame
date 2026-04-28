package contracts

import "github.com/jcsvwinston/GoFrame/pkg/openapi"

func init() {
	RegisterContract(RegisterCommentContract)
}

func RegisterCommentContract(doc *openapi.Document) {
	doc.AddSchema("CommentRecord", openapi.ObjectSchema(map[string]openapi.Schema{
		"id":   openapi.IDSchema(),
		"name": {Type: "string"},
	}, "id", "name"))

	doc.AddSchema("CreateCommentInput", openapi.ObjectSchema(map[string]openapi.Schema{
		"name": {Type: "string"},
	}, "name"))

	doc.AddSchema("UpdateCommentInput", openapi.ObjectSchema(map[string]openapi.Schema{
		"name": {Type: "string"},
	}, "name"))

	doc.EnsurePaths()
	doc.Paths["/comments"] = openapi.PathItem{
		Get: &openapi.Operation{
			OperationID: "listComments",
			Summary:     "List Comments",
			Description: "Returns the scaffolded comments collection.",
			Tags:        []string{"comments"},
			Parameters: []openapi.Parameter{
				openapi.SearchQueryParameter("Filter comments by name."),
			},
			Responses: map[string]openapi.Response{
				"200": openapi.JSONResponse("Resource collection", openapi.CollectionEnvelopeSchema(openapi.RefSchema("CommentRecord"))),
				"500": openapi.ErrorResponse("Unexpected error"),
			},
		},
		Post: &openapi.Operation{
			OperationID: "createComment",
			Summary:     "Create Comment",
			Description: "Creates a scaffolded comments resource.",
			Tags:        []string{"comments"},
			RequestBody: openapi.JSONRequestBody(openapi.RefSchema("CreateCommentInput"), true),
			Responses: map[string]openapi.Response{
				"201": openapi.JSONResponse("Created resource", openapi.DataEnvelopeSchema(openapi.RefSchema("CommentRecord"))),
				"400": openapi.ErrorResponse("Invalid request"),
			},
		},
	}

	doc.Paths["/comments/{id}"] = openapi.PathItem{
		Get: &openapi.Operation{
			OperationID: "getComment",
			Summary:     "Get Comment",
			Description: "Returns one scaffolded Comment resource by id.",
			Tags:        []string{"comments"},
			Parameters: []openapi.Parameter{
				openapi.PathParameter("id", openapi.IDSchema(), "Comment identifier"),
			},
			Responses: map[string]openapi.Response{
				"200": openapi.JSONResponse("Single resource", openapi.DataEnvelopeSchema(openapi.RefSchema("CommentRecord"))),
				"400": openapi.ErrorResponse("Invalid request"),
				"404": openapi.ErrorResponse("Resource not found"),
			},
		},
		Put: &openapi.Operation{
			OperationID: "updateComment",
			Summary:     "Update Comment",
			Description: "Updates one scaffolded Comment resource by id.",
			Tags:        []string{"comments"},
			Parameters: []openapi.Parameter{
				openapi.PathParameter("id", openapi.IDSchema(), "Comment identifier"),
			},
			RequestBody: openapi.JSONRequestBody(openapi.RefSchema("UpdateCommentInput"), true),
			Responses: map[string]openapi.Response{
				"200": openapi.JSONResponse("Updated resource", openapi.DataEnvelopeSchema(openapi.RefSchema("CommentRecord"))),
				"400": openapi.ErrorResponse("Invalid request"),
				"404": openapi.ErrorResponse("Resource not found"),
			},
		},
		Delete: &openapi.Operation{
			OperationID: "deleteComment",
			Summary:     "Delete Comment",
			Description: "Deletes one scaffolded Comment resource by id.",
			Tags:        []string{"comments"},
			Parameters: []openapi.Parameter{
				openapi.PathParameter("id", openapi.IDSchema(), "Comment identifier"),
			},
			Responses: map[string]openapi.Response{
				"204": openapi.EmptyResponse("Resource deleted"),
				"400": openapi.ErrorResponse("Invalid request"),
				"404": openapi.ErrorResponse("Resource not found"),
			},
		},
	}
}
