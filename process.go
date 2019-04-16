
package errands

import (
	"fmt"
	time "time"
	// "sync/atomic"
	// list "container/list"
	schemas "github.com/polygon-io/errands-server/schemas"
)




type Processor struct {
	Parent 			*ErrandsAPI
	Topic 			string
	Concurrency 	int
	Paused 			bool
	Quit 			chan int
	ErrandQueue 	chan *schemas.Errand
	// List 			*list.List
	Fn 				func( *schemas.Errand ) ( map[string]interface{}, error )
	Procs 			[]*ProcThread
}


func ( e *ErrandsAPI ) NewProcessor(
	topic string, concurrency int, 
	fn func( *schemas.Errand ) ( map[string]interface{}, error ) ) ( *Processor, error ) {
	// Create the processor:
	obj := &Processor{}
	obj.Parent = e
	obj.Topic = topic
	obj.Concurrency = concurrency
	obj.Fn = fn
	obj.Paused = false
	obj.Quit = make(chan int)
	obj.ErrandQueue = make( chan *schemas.Errand )
	// Add it to this APIs Processor list:
	e.Processors = append( e.Processors, obj )
	// Actually run the processor:
	go obj.Run()
	return obj, nil

}

func ( p *Processor) Pause(){
	p.Paused = true
}
func ( p *Processor) Resume(){
	p.Paused = false
}

func ( p *Processor ) requestErrandToProcess(){
	errandRes, err := p.Parent.RequestErrandToProcess( p.Topic ); if err != nil {
		fmt.Println("Error requesting errand to process:", err)
		return
	}
	if errandRes.Results.ID != "" {
		p.ErrandQueue <- &errandRes.Results
	}
}


func ( p *Processor ) Run(){
	ticker := time.NewTicker( 4 * time.Second )
	// Create the actually processor threads:
	for i := 1; i <= p.Concurrency; i++ {
		obj := p.NewProcThread()
		p.Procs = append( p.Procs, obj )
		go obj.RunThread()
	}
	// Every so often, if proc threads are awaiting jobs, request them:
	for {
	   select {
		case <- ticker.C:
			// We have room for another item:
			if p.procsAwaitingErrands() && len( p.ErrandQueue ) == 0 && !p.Paused {
				p.requestErrandToProcess()
			}
		case <- p.Quit:
			ticker.Stop()
			return
		}
	}
}


func ( p *Processor ) procsAwaitingErrands() bool {
	for _, proc := range p.Procs {
		if proc.AwaitingErrand {
			return true
		}
	}
	return false
}






type ProcThread struct {
	Processor 			*Processor
	AwaitingErrand 		bool
}


func ( p *Processor ) NewProcThread() *ProcThread {
	obj := &ProcThread{}
	obj.AwaitingErrand = true
	obj.Processor = p
	return obj
}


// Run the actual processor function on items:
func ( proc *ProcThread ) RunThread(){
	for {
		select {
		case job := <- proc.Processor.ErrandQueue:
			// Actually Processing the job:
			res, err := proc.Processor.Fn( job )
			if err != nil {
				fmt.Println("Error processing:", job.ID, "Err:", err)
				proc.Processor.Parent.FailErrand( job.ID, err.Error() )
			}else{
				fmt.Println("Completed Processing:", job.ID)
				proc.Processor.Parent.CompleteErrand( job.ID, res )
			}
			proc.AwaitingErrand = true
		}
	}
}







