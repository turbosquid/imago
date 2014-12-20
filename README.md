## Imago
My first attempt at building somethng in Go. Imago is a simple web front-end to imagemagick to allow mostly 
arbitrary image file conversions. Not pretty or very secure, Imago is realy just a vehicle for me to
familiarize myself with the language. Work requests are POSTed via the API as json objects, and a 
token is returned to the caller which can be used to check the job status. The conversion is executed asychronously 
via a pool of goroutine workers.

## API
Imago implements a simple RESTful API

### Requesting an image conversion
You may request one or more image conversions by passing in an array of actions:

    {
      "actions" : [
        {
          "infile": "s3://mybucket/imago/foo.png",
          "outfile": "s3://mybucket/imago/out/foo 200x200.jpg",
          "operations" : ["resize 200X200", "quality 100"],
          "mimetype":  "image/jpg"
        },
        {
          "infile": "s3://mybucket/imago/foo.png",
          "outfile": "s3://mybucket/imago/out/foo-4xx4k.tiff",
          "operations" : ["resize 4000X4000"],
          "mimetype":  "image/tiff"
        }
      ]
    }

`infile` and `outfile` point to files in s3 to convert. The files will be downloaded and converted using the imagemagick operations specidied in `operations` and reuploaded with the specified mimetype. Note that we use etags
to ensure that we don't needlessly redownload files in a single session.For now, only the `s3` remote file type is supported; other remote filesystem support is coming.

Once a job is submitted, a response is returned with the work id and current queue size (the number of jobs awaiting processing on the queue):

    {
       "id": "f2fcbe5fd9754e80b4d90b509ea42168",
       "queue_length": 3,
       "status": "ok"
    }

### Checking the conversion progress    
You may then fetch the status of your job with a GET request to:

    /api/v1/work/f2fcbe5fd9754e80b4d90b509ea42168
    
You can long-poll, by adding a timeout (in seconds) as a query parameter

    /api/v1/work/f2fcbe5fd9754e80b4d90b509ea42168?timeout=300 
    
In each case, a JSON package is returned with the status of your work request:

    {
        "id": "e2ffd1057f2d46a9a5d599d8379741e0",
        "status": "done",
        "actions": [
            {
                "status": "done",
                "infile": "s3://mybucket/imago/foo.png",
                "outfile": "s3://mybucket/imago/out/foo 200x200.jpg",
                "mimetype": "image/jpg",
                "operations": [
                    "resize 200X200",
                    "quality 100"
                ],
                "output": "",
                "error": ""
            },
            {
                "status": "done",
                "infile": "s3://mybucket/imago/foo.png",
                "outfile": "s3://mybucket/imago/out/foo-4xx4k.tiff",
                "mimetype": "image/tiff",
                "operations": [
                    "resize 4000X4000"
                ],
                "output": "",
                "error": ""
            }
        ]
    }
    
