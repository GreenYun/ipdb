// Package ipdb provides facilities for parsing IPDB files served by ipip.net,
// (maybe) works better than the official package. Compared to the official
// implementation, it provides some easier ways to look into the database,
// and allows you to focus more on your project without hesitation.
//
// A simple way to make queries to the database is, of course the database must
// be loaded in ``IPDb'', Search the IP, then pass the return value as offset
// and Get what you want.
//
// Besides, the implementation of iptree (type Node and Leaf) is to help buiding
// the full database in to an binary tree, maybe useful in some particular
// situations.
package ipdb
