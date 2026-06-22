//go:build windows

package trash

import (
	"fmt"
	"path/filepath"
	"strings"
	"syscall"
	"unicode/utf16"
	"unsafe"
)

const (
	foDelete          = 0x0003
	fofNoConfirmation = 0x0010
	fofAllowUndo      = 0x0040
	fofNoErrorUI      = 0x0400
)

type shFileOpStructW struct {
	hwnd                  uintptr
	wFunc                 uint32
	pFrom                 *uint16
	pTo                   *uint16
	fFlags                uint16
	fAnyOperationsAborted int32
	hNameMappings         uintptr
	lpszProgressTitle     *uint16
}

var (
	shell32          = syscall.NewLazyDLL("shell32.dll")
	shFileOperationW = shell32.NewProc("SHFileOperationW")
)

func MoveToTrash(paths ...string) error {
	for _, path := range paths {
		abs, err := filepath.Abs(path)
		if err != nil {
			return fmt.Errorf("resolve %q: %w", path, err)
		}

		// SHFileOperation does not support extended-length \\?\ paths.
		if strings.HasPrefix(abs, `\\?\`) {
			return fmt.Errorf(
				"cannot send extended-length path to Recycle Bin: %q",
				abs,
			)
		}

		// Windows requires a double-NUL-terminated path list.
		from := utf16.Encode([]rune(abs))
		from = append(from, 0, 0)

		op := shFileOpStructW{
			wFunc: foDelete,
			pFrom: &from[0],
			fFlags: fofAllowUndo |
				fofNoConfirmation |
				fofNoErrorUI,
		}

		result, _, _ := shFileOperationW.Call(
			uintptr(unsafe.Pointer(&op)),
		)

		if result != 0 {
			return fmt.Errorf(
				"send %q to Recycle Bin: SHFileOperationW returned %#x",
				abs,
				result,
			)
		}

		if op.fAnyOperationsAborted != 0 {
			return fmt.Errorf(
				"send %q to Recycle Bin: operation cancelled",
				abs,
			)
		}
	}

	return nil
}
