// Package admin implements the sicau-niu operator-facing HTTP controllers for
// the LinaPro management console: college dictionary CRUD and a read-only player
// query. All endpoints are protected by the host unified
// Auth+Tenancy+Permission middleware chain, with the concrete permission
// identifier declared on each API DTO g.Meta tag (sicau-niu:college:* and
// sicau-niu:player:list).
package admin
