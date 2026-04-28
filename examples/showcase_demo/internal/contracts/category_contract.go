package contracts

import "github.com/jcsvwinston/GoFrame/pkg/openapi"

func init() {
	RegisterContract(RegisterCategoryContract)
}

func RegisterCategoryContract(doc *openapi.Document) {
	doc.AddSchema("CategoryRecord", openapi.ObjectSchema(map[string]openapi.Schema{
		"id":   openapi.IDSchema(),
		"name": {Type: "string"},
	}, "id", "name"))

	doc.AddSchema("CreateCategoryInput", openapi.ObjectSchema(map[string]openapi.Schema{
		"name": {Type: "string"},
	}, "name"))

	doc.AddSchema("UpdateCategoryInput", openapi.ObjectSchema(map[string]openapi.Schema{
		"name": {Type: "string"},
	}, "name"))

	doc.EnsurePaths()
	doc.Paths["/categories"] = openapi.PathItem{
		Get: &openapi.Operation{
			OperationID: "listCategories",
			Summary:     "List Categories",
			Description: "Returns the scaffolded categories collection.",
			Tags:        []string{"categories"},
			Parameters: []openapi.Parameter{
				openapi.SearchQueryParameter("Filter categories by name."),
			},
			Responses: map[string]openapi.Response{
				"200": openapi.JSONResponse("Resource collection", openapi.CollectionEnvelopeSchema(openapi.RefSchema("CategoryRecord"))),
				"500": openapi.ErrorResponse("Unexpected error"),
			},
		},
		Post: &openapi.Operation{
			OperationID: "createCategory",
			Summary:     "Create Category",
			Description: "Creates a scaffolded categories resource.",
			Tags:        []string{"categories"},
			RequestBody: openapi.JSONRequestBody(openapi.RefSchema("CreateCategoryInput"), true),
			Responses: map[string]openapi.Response{
				"201": openapi.JSONResponse("Created resource", openapi.DataEnvelopeSchema(openapi.RefSchema("CategoryRecord"))),
				"400": openapi.ErrorResponse("Invalid request"),
			},
		},
	}

	doc.Paths["/categories/{id}"] = openapi.PathItem{
		Get: &openapi.Operation{
			OperationID: "getCategory",
			Summary:     "Get Category",
			Description: "Returns one scaffolded Category resource by id.",
			Tags:        []string{"categories"},
			Parameters: []openapi.Parameter{
				openapi.PathParameter("id", openapi.IDSchema(), "Category identifier"),
			},
			Responses: map[string]openapi.Response{
				"200": openapi.JSONResponse("Single resource", openapi.DataEnvelopeSchema(openapi.RefSchema("CategoryRecord"))),
				"400": openapi.ErrorResponse("Invalid request"),
				"404": openapi.ErrorResponse("Resource not found"),
			},
		},
		Put: &openapi.Operation{
			OperationID: "updateCategory",
			Summary:     "Update Category",
			Description: "Updates one scaffolded Category resource by id.",
			Tags:        []string{"categories"},
			Parameters: []openapi.Parameter{
				openapi.PathParameter("id", openapi.IDSchema(), "Category identifier"),
			},
			RequestBody: openapi.JSONRequestBody(openapi.RefSchema("UpdateCategoryInput"), true),
			Responses: map[string]openapi.Response{
				"200": openapi.JSONResponse("Updated resource", openapi.DataEnvelopeSchema(openapi.RefSchema("CategoryRecord"))),
				"400": openapi.ErrorResponse("Invalid request"),
				"404": openapi.ErrorResponse("Resource not found"),
			},
		},
		Delete: &openapi.Operation{
			OperationID: "deleteCategory",
			Summary:     "Delete Category",
			Description: "Deletes one scaffolded Category resource by id.",
			Tags:        []string{"categories"},
			Parameters: []openapi.Parameter{
				openapi.PathParameter("id", openapi.IDSchema(), "Category identifier"),
			},
			Responses: map[string]openapi.Response{
				"204": openapi.EmptyResponse("Resource deleted"),
				"400": openapi.ErrorResponse("Invalid request"),
				"404": openapi.ErrorResponse("Resource not found"),
			},
		},
	}
}
