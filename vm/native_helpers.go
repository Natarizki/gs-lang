package vm

import (
	"fmt"
	"math/rand"
	"os"
	"time"
)

func osReadFile(path string) ([]byte, error) {
	return os.ReadFile(path)
}

func osWriteFile(path string, data []byte) error {
	return os.WriteFile(path, data, 0644)
}

func osAppendFile(path string, content string) error {
	f, err := os.OpenFile(path, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		return err
	}
	defer f.Close()
	_, err = f.WriteString(content)
	return err
}

func osRemoveFile(path string) error {
	return os.Remove(path)
}

func nowString() string {
	return time.Now().Format(time.RFC3339)
}

func sleepMillis(ms int64) {
	time.Sleep(time.Duration(ms) * time.Millisecond)
}

func formatTimeNow(layout string) string {
	// konversi layout sederhana gaya GS jadi layout Go, contoh dasar
	goLayout := "2006-01-02 15:04:05"
	if layout != "" {
		goLayout = layout
	}
	return time.Now().Format(goLayout)
}

func randomInRange(min, max int64) int64 {
	if max <= min {
		return min
	}
	return min + rand.Int63n(max-min+1)
}

func exitProgram(code int) {
	fmt.Println()
	os.Exit(code)
}

// deepCopyValue membuat salinan mendalam dari sebuah Value (untuk List dan Map)
func deepCopyValue(v Value) Value {
	switch val := v.(type) {
	case *List:
		newElements := make([]Value, len(val.Elements))
		for i, el := range val.Elements {
			newElements[i] = deepCopyValue(el)
		}
		return &List{Elements: newElements}
	case *Map:
		newMap := NewMap()
		for _, pair := range val.Pairs {
			newMap.Set(deepCopyValue(pair.Key), deepCopyValue(pair.Value))
		}
		return newMap
	case *Int:
		return &Int{Value: val.Value}
	case *Float:
		return &Float{Value: val.Value}
	case *String:
		return &String{Value: val.Value}
	case *Bool:
		return &Bool{Value: val.Value}
	default:
		return v
	}
}
