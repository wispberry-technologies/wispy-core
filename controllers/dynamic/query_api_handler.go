// Indent to build a dynamic query API handler for the application.
// ```
// {
//   "fetch-projects":{
//     "method": "GET",
//     "url": "/api/v1/projects",
//     "description": "Fetch all projects for the authenticated user.",
//     "handler": "dynamicQueryApiHandler",
//     "protected": {
//       "type": "user",
//       "roles": [],
//       "sql_column_match": "user_id"
//     },
//     "queryParams": {
//       "status": {
//         "type": "string",
//         "description": "Filter projects by status (e.g., active, archived).",
//         "required": false
//       },
//       "sortBy": {
//         "type": "string",
//         "description": "Sort projects by a specific field (e.g., name, createdAt).",
//         "required": false
//       },
//       "limit": {
//         "type": "integer",
//         "description": "Limit the number of projects returned.",
//         "required": false
//       },
//       "offset": {
//         "type": "integer",
//         "description": "Offset for pagination.",
//         "required": false
//       }
//     }
//   }
// }
// ```

package controllers

import (
	"encoding/json"
	"net/http"
)

func dynamicQueryApiHandler(w http.ResponseWriter, r *http.Request) {
	// Extract the query parameters from the request
	queryParams := r.URL.Query()

	// Build the SQL query dynamically based on the provided parameters
	sqlQuery := "SELECT * FROM projects WHERE 1=1"
	var args []interface{}

	// Example: Add filters based on query parameters
	if status := queryParams.Get("status"); status != "" {
		sqlQuery += " AND status = ?"
		args = append(args, status)
	}

	if sortBy := queryParams.Get("sortBy"); sortBy != "" {
		sqlQuery += " ORDER BY " + sortBy
	}

	if limit := queryParams.Get("limit"); limit != "" {
		sqlQuery += " LIMIT ?"
		args = append(args, limit)
	}

	if offset := queryParams.Get("offset"); offset != "" {
		sqlQuery += " OFFSET ?"
		args = append(args, offset)
	}

	// Execute the SQL query and return the results
	results, err := ExecuteSQLQuery(db, sqlQuery, args...)
	if err != nil {
		http.Error(w, "Failed to execute query: "+err.Error(), http.StatusInternalServerError)
		return
	}

	// Return the results as JSON
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(results)
}
