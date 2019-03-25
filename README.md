# Go client for Errands API
Golang client for Errands API server.

This only has simple update methods and processing functionality. 

### Creating new instance:

```golang
// Create new API:
api := New("http://localhost:5555")
```

### Fetching from API:
```golang
errands, _ := api.GetErrands()
fmt.Println( "Got Errands:", errands.Results )
```

### Processing

Each processing function will be executed in it's own gorouting. Once completed it will wait for another errand to process. The errand will be marked as failed/completed depending on the returned values. 

You can also use `processor.Pause()` and `processor.Resume()`. However, note that the processors that are currently processing when you call Pause, will not stop. This only prevents future errands from being processed.

```golang
/* 
Process to run for every errand:
- If you return an Error, the errand will automatically be failed with the reason
as the error returnd.
- If you return no error, the results will be sent and the errand will automatically
be marked as completed
*/
fn := func( errand *schemas.Errand ) ( map[string]interface{}, error ){
	fmt.Println("Processing:", errand.Name, " - ID: ", errand.ID)
	return map[string]interface{}{ "results": "OK", }, nil
}
// Parameters: ( Errand Type, Concurrency, Func )
processor, _ := api.NewProcessor( "tester", 1, fn )
```