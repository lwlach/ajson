package ajson

import "io"

func Unmarshal(body []byte, clone bool) (root *Node, err error) {
	buf := newBuffer(body, clone)
	var (
		b       byte
		index   *int
		key     []byte
		current *Node
	)
	// main loop: detect all parts of json struct
base:
	for {
		// detect type of element
		b, err = buf.first()
		if err != nil {
			break
		}
		switch true {
		case b == bracesL:
			// Detected: Object
		case b == bracketL:
			// Detected: Array
		case b == quotes:
			// Detected: String
			current = newNode(current, buf, String, key, index)
			err = buf.string()
			current.borders[1] = buf.index + 1
			if err == nil {
				err = buf.step()
			}
			if err != nil {
				break base
			}
		case (b >= '0' && b <= '9') || b == '.' || b == '+' || b == '-' || b == 'e' || b == 'E':
			// Detected: Numeric
			current = newNode(current, buf, Numeric, key, index)
			err = buf.numeric()
			current.borders[1] = buf.index
			if err != nil {
				break base
			}
		case b == 'n' || b == 'N':
			// mb: Null
			current = newNode(current, buf, Null, key, index)
			err = buf.null()
			current.borders[1] = buf.index + 1
			if err == nil {
				err = buf.step()
			}
			if err != nil {
				break base
			}
		case b == 't' || b == 'T' || b == 'f' || b == 'F':
			// mb: Bool
			current = newNode(current, buf, Bool, key, index)
			if b == 't' || b == 'T' {
				err = buf.true()
			} else {
				err = buf.false()
			}
			current.borders[1] = buf.index + 1
			if err == nil {
				err = buf.step()
			}
			if err != nil {
				break base
			}
		default:
			return nil, errorSymbol(b, buf.index)
		}
	}

	// outer
	if err == io.EOF {
		err = nil
		root = current
	}

	return
}
