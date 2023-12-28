function onTrafficFromClient(ctx, req) {
  const msg = req.Fields["request"].GetBytesValue()
  switch (String.fromCharCode(msg[0])) {
    case "Q":
      // The client is asking for a query
      const query = String.fromCharCode(...msg.slice(5, -1))
      console.log("query:", query)
      // Parse the query and log it to the console
      // The parseSQL function returns a stringified JSON object
      const parsedQuery = parseSQL(query)
      console.log("parsed query:", parsedQuery)
      break
  }

  // Terminate the request immediately by modifying the request object
  // Value is a helper function to create a value object in JS
  // FIXME: The response is not working yet, but the terminate is
  // req.Fields["response"] = Value(bytes("Hello from JS"))
  // req.Fields["terminate"] = Value(true)

  // Log the request to the console, which will be visible in the
  // gatewayd logs
  // console.log("onTrafficFromClient is called from JS", req)

  // Return the (modified) request object
  return req
}
