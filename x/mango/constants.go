package mango

type operator string

const (
	opNone = operator("")

	// Combination operators - http://docs.couchdb.org/en/2.0.0/api/database/find.html#combination-operators
	opAnd = operator("$and")
	// opOr        = operator("$or")
	// opNot       = operator("$not")
	// opNor       = operator("$nor")
	// opAll       = operator("$all")
	// opElemMatch = operator("$elemMatch")

	// Condition operators - http://docs.couchdb.org/en/2.0.0/api/database/find.html#condition-operators
	opLT  = operator("$lt")
	opLTE = operator("$lte")
	opEq  = operator("$eq")
	opNE  = operator("$ne")
	opGTE = operator("$gte")
	opGT  = operator("$gt")
	// opExists = operator("$exists")
	// opType   = operator("$type")
	// opIn     = operator("$in")
	// opNIn    = operator("$nin")
	// opSize   = operator("$size")
	// opMod    = operator("$mod")
	// opRegex  = operator("$regex")
)
