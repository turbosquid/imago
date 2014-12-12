## Imago
My first attempt at building somethng in Go. Imago is a simple web front-end to imagemagick to allow mostly 
arbitrary image file conversions. Not pretty or very secure, Imago is realy just a vehicle for me to
familiarize myself with the language. Work requests are POSTed via the API as json objects, and a 
token is returned to the caller which can be used to check the job status. The conversion is executed asychronously 
via a pool of goroutine workers.
