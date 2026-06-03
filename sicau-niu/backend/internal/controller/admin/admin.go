// Package admin implements the sicau-niu operator-facing HTTP controllers for
// the LinaPro management console: college dictionary CRUD, a read-only player
// query, and the C2 content-asset CRUD for cattle, iron-cows, cards and quotes.
// All endpoints are protected by the host unified Auth+Tenancy+Permission
// middleware chain, with the concrete permission identifier declared on each API
// DTO g.Meta tag (sicau-niu:college:*, sicau-niu:player:list, sicau-niu:niu:*,
// sicau-niu:iron:*, sicau-niu:card:* and sicau-niu:quote:*).
package admin
