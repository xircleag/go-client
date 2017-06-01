/*

Package client provides an interface to the Layer Client API (CAPI), handling
underlying authentication and providing high level structures for all major
functions.

More information on the Layer Client API can be found at
http;//docs.layer.com

Authentication

A Layer account is required to use this package.  A number of authentication
methods are supported.  See the examples for standard usage.

Iterable Objects

Certain returned values are iterable.  Examples of this are conversation and
message lists.  This package implements a special that signals the end of an
iterable result.

Contexts

This package implements contexts as described at
https://blog.golang.org/context.
Most functions accept a context, which will be honored for the scope of that
particular function call.

	messages, err := convo.Messages(ctx)
	if err != nil {
		// Handle error
	}
	for {
		message, err := messages.Next()
		if err == iterator.Done {
			// This is the last message
			break
		}
		if err != nil {
			// Handle error
		}

		// Do something with the message
		_ = message
	}


*/
package client
