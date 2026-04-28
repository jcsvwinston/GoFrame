package contracts

import "github.com/jcsvwinston/GoFrame/pkg/openapi"

func init() {
	RegisterContract(RegisterAuthorContract)
}

func RegisterAuthorContract(doc *openapi.Document) {
	doc.AddSchema("AuthorRecord", openapi.ObjectSchema(map[string]openapi.Schema{
		"id":   openapi.IDSchema(),
		"name": {Type: "string"},
	}, "id", "name"))

	doc.AddSchema("CreateAuthorInput", openapi.ObjectSchema(map[string]openapi.Schema{
		"name": {Type: "string"},
	}, "name"))

	doc.AddSchema("UpdateAuthorInput", openapi.ObjectSchema(map[string]openapi.Schema{
		"name": {Type: "string"},
	}, "name"))

	doc.EnsurePaths()
	doc.Paths["/authors"] = openapi.PathItem{
		Get: &openapi.Operation{
			OperationID: "listAuthors",
			Summary:     "List Authors",
			Description: "Returns the scaffolded authors collection.",
			Tags:        []string{"authors"},
			Parameters: []openapi.Parameter{
				openapi.SearchQueryParameter("Filter authors by name."),
			},
			Responses: map[string]openapi.Response{
				"200": openapi.JSONResponse("Resource collection", openapi.CollectionEnvelopeSchema(openapi.RefSchema("AuthorRecord"))),
				"500": openapi.ErrorResponse("Unexpected error"),
			},
		},
		Post: &openapi.Operation{
			OperationID: "createAuthor",
			Summary:     "Create Author",
			Description: "Creates a scaffolded authors resource.",
			Tags:        []string{"authors"},
			RequestBody: openapi.JSONRequestBody(openapi.RefSchema("CreateAuthorInput"), true),
			Responses: map[string]openapi.Response{
				"201": openapi.JSONResponse("Created resource", openapi.DataEnvelopeSchema(openapi.RefSchema("AuthorRecord"))),
				"400": openapi.ErrorResponse("Invalid request"),
			},
		},
	}

	doc.Paths["/authors/{id}"] = openapi.PathItem{
		Get: &openapi.Operation{
			OperationID: "getAuthor",
			Summary:     "Get Author",
			Description: "Returns one scaffolded Author resource by id.",
			Tags:        []string{"authors"},
			Parameters: []openapi.Parameter{
				openapi.PathParameter("id", openapi.IDSchema(), "Author identifier"),
			},
			Responses: map[string]openapi.Response{
				"200": openapi.JSONResponse("Single resource", openapi.DataEnvelopeSchema(openapi.RefSchema("AuthorRecord"))),
				"400": openapi.ErrorResponse("Invalid request"),
				"404": openapi.ErrorResponse("Resource not found"),
			},
		},
		Put: &openapi.Operation{
			OperationID: "updateAuthor",
			Summary:     "Update Author",
			Description: "Updates one scaffolded Author resource by id.",
			Tags:        []string{"authors"},
			Parameters: []openapi.Parameter{
				openapi.PathParameter("id", openapi.IDSchema(), "Author identifier"),
			},
			RequestBody: openapi.JSONRequestBody(openapi.RefSchema("UpdateAuthorInput"), true),
			Responses: map[string]openapi.Response{
				"200": openapi.JSONResponse("Updated resource", openapi.DataEnvelopeSchema(openapi.RefSchema("AuthorRecord"))),
				"400": openapi.ErrorResponse("Invalid request"),
				"404": openapi.ErrorResponse("Resource not found"),
			},
		},
		Delete: &openapi.Operation{
			OperationID: "deleteAuthor",
			Summary:     "Delete Author",
			Description: "Deletes one scaffolded Author resource by id.",
			Tags:        []string{"authors"},
			Parameters: []openapi.Parameter{
				openapi.PathParameter("id", openapi.IDSchema(), "Author identifier"),
			},
			Responses: map[string]openapi.Response{
				"204": openapi.EmptyResponse("Resource deleted"),
				"400": openapi.ErrorResponse("Invalid request"),
				"404": openapi.ErrorResponse("Resource not found"),
			},
		},
	}
}
