package errands

import (
	"fmt"
	time "time"

	schemas "github.com/polygon-io/errands-server/schemas"
)

// Processor is the main struct which handles all the processing for a client
type Processor struct {
	Parent      *ErrandsAPI
	Topic       string
	Concurrency int
	Paused      bool
	Quit        chan int
	ErrandQueue chan *schemas.Errand
	Fn          func(*schemas.Errand) (map[string]interface{}, error)
	Procs       []*ProcThread
}

// NewProcessor creates and returns a *Processor with the params sent.
func (e *ErrandsAPI) NewProcessor(
	topic string, concurrency int,
	fn func(*schemas.Errand) (map[string]interface{}, error)) (*Processor, error) {
	// Create the processor:
	obj := &Processor{
		Parent:      e,
		Topic:       topic,
		Concurrency: concurrency,
		Fn:          fn,
		Paused:      false,
		Quit:        make(chan (int)),
		ErrandQueue: make(chan (*schemas.Errand)),
	}
	// Add it to this APIs Processor list:
	e.Processors = append(e.Processors, obj)
	// Actually run the processor:
	go obj.Run()
	return obj, nil

}

// Pause pauses the processor. This will not pause the current threads, it will
// simply stop the processor from processing subsequent items.
func (p *Processor) Pause() {
	p.Paused = true
}

// Resume tells the processor that it should start processing items again.
func (p *Processor) Resume() {
	p.Paused = false
}

func (p *Processor) requestErrandToProcess() {
	errandRes, err := p.Parent.RequestErrandToProcess(p.Topic)
	if err != nil {
		fmt.Println("Error requesting errand to process:", err)
		return
	}
	if errandRes.Results.ID != "" {
		p.ErrandQueue <- &errandRes.Results
	}
}

// Run creates the threads, and starts the loop to query for jobs to run.
func (p *Processor) Run() {
	ticker := time.NewTicker(4 * time.Second)
	// Create the actually processor threads:
	for i := 1; i <= p.Concurrency; i++ {
		obj := p.NewProcThread()
		p.Procs = append(p.Procs, obj)
		go obj.RunThread()
	}
	// Every so often, if proc threads are awaiting jobs, request them:
	for {
		select {
		case <-ticker.C:
			// We have room for another item:
			if p.procsAwaitingErrands() && len(p.ErrandQueue) == 0 && !p.Paused {
				p.requestErrandToProcess()
			}
		case <-p.Quit:
			ticker.Stop()
			return
		}
	}
}

func (p *Processor) procsAwaitingErrands() bool {
	for _, proc := range p.Procs {
		if proc.AwaitingErrand {
			return true
		}
	}
	return false
}

// ProcThread is created per concurrency. So each actual item processed
// will be inside of a ProcThread.
type ProcThread struct {
	Processor      *Processor
	AwaitingErrand bool
}

// NewProcThread creates and returns a *ProcThread
func (p *Processor) NewProcThread() *ProcThread {
	obj := &ProcThread{
		AwaitingErrand: true,
		Processor:      p,
	}
	return obj
}

// RunThread runs the actual processor function on items:
func (proc *ProcThread) RunThread() {
	for {
		select {
		case job := <-proc.Processor.ErrandQueue:
			proc.AwaitingErrand = false
			// Actually Processing the job:
			res, err := proc.Processor.Fn(job)
			if err != nil {
				fmt.Println("Error processing:", job.ID, "Err:", err)
				proc.Processor.Parent.FailErrand(job.ID, err.Error())
			} else {
				fmt.Println("Completed Processing:", job.ID)
				proc.Processor.Parent.CompleteErrand(job.ID, res)
			}
			proc.AwaitingErrand = true
		}
	}
}
