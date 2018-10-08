package az

// resultPage is the common result interface implemented by paginated APIs.
type resultPage interface {
	Next() error
	NotDone() bool
}

// Paginator simplifies API call pagination. Example usage:
//
//	rp, err := client.ApiCall()
//	pg := az.Paginate(&rp, err)
//	for pg.Next() {
//		// Use rp.Response() or rp.Values()
//	}
//	if pg.Err != nil {
//		// Handler error
//	}
type Paginator struct {
	Err error

	rp   resultPage
	next bool
}

// Paginate returns a simple interface for iterating over all API result pages.
func Paginate(rp resultPage, err error) Paginator {
	if err != nil || !rp.NotDone() {
		rp = nil
	}
	return Paginator{Err: err, rp: rp}
}

// Next returns true while there is more data available.
func (p *Paginator) Next() bool {
	if p.rp != nil {
		if !p.next {
			p.next = true
			return true
		}
		if p.Err = p.rp.Next(); p.Err == nil && p.rp.NotDone() {
			return true
		}
		p.rp = nil
	}
	return false
}
