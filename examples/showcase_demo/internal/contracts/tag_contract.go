package contracts

import "github.com/jcsvwinston/GoFrame/pkg/openapi"

func init() {
	RegisterContract(RegisterTagContract)
}

func RegisterTagContract(doc *openapi.Document) {
	doc.AddSchema("TagRecord", openapi.ObjectSchema(map[string]openapi.Schema{
		"id":   openapi.IDSchema(),
		"name": {Type: "string"},
	}, "id", "name"))

	doc.AddSchema("CreateTagInput", openapi.ObjectSchema(map[string]openapi.Schema{
		"name": {Type: "string"},
	}, "name"))

	doc.AddSchema("UpdateTagInput", openapi.ObjectSchema(map[string]openapi.Schema{
		"name": {Type: "string"},
	}, "name"))

	doc.EnsurePaths()
	doc.Paths["/tags"] = openapi.PathItem{
		Get: &openapi.Operation{
			OperationID: "listTags",
			Summary:     "List Tags",
			Description: "Returns the scaffolded tags collection.",
			Tags:        []string{"tags"},
			Parameters: []openapi.Parameter{
				openapi.SearchQueryParameter("Filter tags by name."),
			},
			Responses: map[string]openapi.Response{
				"200": openapi.JSONResponse("Resource collection", openapi.CollectionEnvelopeSchema(openapi.RefSchema("TagRecord"))),
				"500": openapi.ErrorResponse("Unexpected error"),
			},
		},
		Post: &openapi.Operation{
			OperationID: "createTag",
			Summary:     "Create Tag",
			Description: "Creates a scaffolded tags resource.",
			Tags:        []string{"tags"},
			RequestBody: openapi.JSONRequestBody(openapi.RefSchema("CreateTagInput"), true),
			Responses: map[string]openapi.Response{
				"201": openapi.JSONResponse("Created resource", openapi.DataEnvelopeSchema(openapi.RefSchema("TagRecord"))),
				"400": openapi.ErrorResponse("Invalid request"),
			},
		},
	}

	doc.Paths["/tags/{id}"] = openapi.PathItem{
		Get: &openapi.Operation{
			OperationID: "getTag",
			Summary:     "Get Tag",
			Description: "Returns one scaffolded Tag resource by id.",
			Tags:        []string{"tags"},
			Parameters: []openapi.Parameter{
				openapi.PathParameter("id", openapi.IDSchema(), "Tag identifier"),
			},
			Responses: map[string]openapi.Response{
				"200": openapi.JSONResponse("Single resource", openapi.DataEnvelopeSchema(openapi.RefSchema("TagRecord"))),
				"400": openapi.ErrorResponse("Invalid request"),
				"404": openapi.ErrorResponse("Resource not found"),
			},
		},
		Put: &openapi.Operation{
			OperationID: "updateTag",
			Summary:     "Update Tag",
			Description: "Updates one scaffolded Tag resource by id.",
			Tags:        []string{"tags"},
			Parameters: []openapi.Parameter{
				openapi.PathParameter("id", openapi.IDSchema(), "Tag identifier"),
			},
			RequestBody: openapi.JSONRequestBody(openapi.RefSchema("UpdateTagInput"), true),
			Responses: map[string]openapi.Response{
				"200": openapi.JSONResponse("Updated resource", openapi.DataEnvelopeSchema(openapi.RefSchema("TagRecord"))),
				"400": openapi.ErrorResponse("Invalid request"),
				"404": openapi.ErrorResponse("Resource not found"),
			},
		},
		Delete: &openapi.Operation{
			OperationID: "deleteTag",
			Summary:     "Delete Tag",
			Description: "Deletes one scaffolded Tag resource by id.",
			Tags:        []string{"tags"},
			Parameters: []openapi.Parameter{
				openapi.PathParameter("id", openapi.IDSchema(), "Tag identifier"),
			},
			Responses: map[string]openapi.Response{
				"204": openapi.EmptyResponse("Resource deleted"),
				"400": openapi.ErrorResponse("Invalid request"),
				"404": openapi.ErrorResponse("Resource not found"),
			},
		},
	}
}
