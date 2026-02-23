//go:build nochdb

package chdb

import "fmt"

func initEmbedded() error {
	return fmt.Errorf("embedded ClickHouse not available — this binary was built with the nochdb tag")
}
