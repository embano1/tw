package main

import "fmt"

func usage() string {
	return fmt.Sprint(`
  ______   __     __    
 /\__  _\ /\ \  _ \ \   
 \/_/\ \/ \ \ \/ ".\ \  
    \ \_\  \ \__/".~\_\ 
     \/_/   \/_/   \/_/ 
						
  A minimal Twitter CLI
	  
  This application uses Oauthv1 to securely authenticate requests.
  You can obtain API credentials from https://apps.twitter.com/.
  Always handle secrets carefully!
  
  If you don't provide a credentials file, please provide credentials via the following environment variables: 
  TW_CONSUMER_KEY, TW_CONSUMER_SECRET, TW_ACCESS_TOKEN, TW_TOKEN_SECRET
		  `)
}
