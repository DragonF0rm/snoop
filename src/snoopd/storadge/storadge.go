package storadge

//func SaveRequest (req http.Request) {
//	var request []string
//	reqStr := fmt.Sprintf("%v %v %v", req.Method, req.URL, req.Proto)
//	request = append(request, reqStr)
//	for n, headers := range req.Header {
//		for _, h := range headers {
//			request = append(request, fmt.Sprintf(“%v: %v”, name, h))
//		}
//	}
//
//	// If this is a POST, add post data
//	if r.Method == “POST” {
//		r.ParseForm()
//		request = append(request, “\n”)
//		request = append(request, r.Form.Encode())
//	}
//	// Return the request as a string
//	return strings.Join(request, “\n”)
//}
