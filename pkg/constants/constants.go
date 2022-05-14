package constants

// Verbs are API request verbs.
var Verbs = []string{CREATE, GET, UPDATE, PATCH, LIST, DELETE, ALL}

const (
	// CREATE is verb for create API requests.
	CREATE string = "create"
	// GET is verb for get API requests.
	GET string = "get"
	// UPDATE is verb for update API requests.
	UPDATE string = "update"
	// PATCH is verb for patch API requests.
	PATCH string = "patch"
	// LIST is verb for list API requests.
	LIST string = "list"
	// DELETE is verb for delete API requests.
	DELETE string = "delete"
	// ALL is verb for all API requests.
	ALL string = "all"
)
