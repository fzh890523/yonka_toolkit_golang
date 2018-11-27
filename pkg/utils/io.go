package utils

import (
	"io"
)

// 这里不做buf，依赖caller自己对reader和writer wrap为buffered
func Copy(r io.Reader, w io.Writer) (read, write int, err error) {
	buf := make([]byte, 1024)
	eof := false
	n := 0
	for !eof {
		n, err = r.Read(buf)
		if err != nil {
			if err == io.EOF {
				eof = true
				err = nil
			}
		}

		read += n

		if err != nil {
			return
		}

		if n > 0 {
			m := 0
			left := n

			for left > 0 {
				m, err = w.Write(buf[n-left:])
				write += m
				if err != nil {
					return
				}
				left -= m
			}
		}
	}

	return
}
