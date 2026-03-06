package sparse

import "github.com/tarcisiozf/wasp/internal/serialization"

func Encode(encoder *serialization.Encoder, mem *Memory) {
	encoder.Int(mem.numPages)
	encoder.Int(mem.maxPages)
	encoder.Int(mem.pageSize)
	encoder.Int(mem.pagesWithData)
	encoder.Float64(mem.mergeThreshold)

	encoder.Int(len(mem.pages)) // total slots (capacity)
	for i, page := range mem.pages {
		if page == nil {
			continue
		}
		encoder.Int(i)          // page index
		encoder.RawBytes(*page) // page data
	}
}

func Decode(decoder *serialization.Decoder) (*Memory, error) {
	numPages, err := decoder.Int()
	if err != nil {
		return nil, err
	}
	maxPages, err := decoder.Int()
	if err != nil {
		return nil, err
	}
	pageSize, err := decoder.Int()
	if err != nil {
		return nil, err
	}
	pagesWithData, err := decoder.Int()
	if err != nil {
		return nil, err
	}
	mergeThreshold, err := decoder.Float64()
	if err != nil {
		return nil, err
	}

	totalSlots, err := decoder.Int()
	if err != nil {
		return nil, err
	}
	pages := make([]*[]byte, totalSlots)
	for i := 0; i < pagesWithData; i++ {
		pageIndex, err := decoder.Int()
		if err != nil {
			return nil, err
		}
		pageData, err := decoder.BytesN(pageSize)
		if err != nil {
			return nil, err
		}
		pages[pageIndex] = &pageData
	}

	return &Memory{
		numPages: numPages,
		maxPages: maxPages,

		pageSize:      pageSize,
		pagesWithData: pagesWithData,
		pages:         pages,

		mergeThreshold: mergeThreshold,
	}, nil
}
