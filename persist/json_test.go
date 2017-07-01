package persist

import (
	"bytes"
	"os"
	"path/filepath"
	"sync"
	"testing"

	"github.com/NebulousLabs/Sia/build"
)

// TestSaveLoadJSON creates a simple object and then tries saving and loading
// it.
func TestSaveLoadJSON(t *testing.T) {
	if testing.Short() {
		t.SkipNow()
	}
	// Create the directory used for testing.
	dir := filepath.Join(build.TempDir(persistDir), t.Name())
	err := os.MkdirAll(dir, 0700)
	if err != nil {
		t.Fatal(err)
	}

	// Create and save the test object.
	testMeta := Metadata{"Test Struct", "v1.2.1"}
	type testStruct struct {
		One   string
		Two   uint64
		Three []byte
	}

	obj1 := testStruct{"dog", 25, []byte("more dog")}
	obj1Filename := filepath.Join(dir, "obj1.json")
	err = SaveJSON(testMeta, obj1, obj1Filename)
	if err != nil {
		t.Fatal(err)
	}
	var obj2 testStruct

	// Try loading the object
	err = LoadJSON(testMeta, &obj2, obj1Filename)
	if err != nil {
		t.Fatal(err)
	}
	// Verify equivalence.
	if obj2.One != obj1.One {
		t.Error("persist mismatch")
	}
	if obj2.Two != obj1.Two {
		t.Error("persist mismatch")
	}
	if !bytes.Equal(obj2.Three, obj1.Three) {
		t.Error("persist mismatch")
	}
	if obj2.One != "dog" {
		t.Error("persist mismatch")
	}
	if obj2.Two != 25 {
		t.Error("persist mismatch")
	}
	if !bytes.Equal(obj2.Three, []byte("more dog")) {
		t.Error("persist mismatch")
	}

	// Try loading the object using the temp file.
	err = LoadJSON(testMeta, &obj2, obj1Filename+tempSuffix)
	if err != ErrBadFilenameSuffix {
		t.Error("did not get bad filename suffix")
	}

	// Try saving the object multiple times concurrently.
	var wg sync.WaitGroup
	errs := make([]bool, 250)
	for i := 0; i < 250; i++ {
		wg.Add(1)
		go func(i int) {
			defer wg.Done()
			defer func() {
				r := recover() // Error is irrelevant, managed by err slice.
				if r != nil {
					errs[i] = true
				}
			}()
			SaveJSON(testMeta, obj1, obj1Filename)
		}(i)
	}
	wg.Wait()
	// At least one of the saves should have complained about concurrent usage.
	var found bool
	for i := range errs {
		if errs[i] {
			found = true
			break
		}
	}
	if !found {
		// Single core machines could result in this error.
		t.Log("File usage overlap detector seems to be ineffective")
	}

	// Despite the errors, the object should still be readable.
	err = LoadJSON(testMeta, &obj2, obj1Filename)
	if err != nil {
		t.Fatal(err)
	}
	// Verify equivalence.
	if obj2.One != obj1.One {
		t.Error("persist mismatch")
	}
	if obj2.Two != obj1.Two {
		t.Error("persist mismatch")
	}
	if !bytes.Equal(obj2.Three, obj1.Three) {
		t.Error("persist mismatch")
	}
	if obj2.One != "dog" {
		t.Error("persist mismatch")
	}
	if obj2.Two != 25 {
		t.Error("persist mismatch")
	}
	if !bytes.Equal(obj2.Three, []byte("more dog")) {
		t.Error("persist mismatch")
	}
}

// TestLoadJSONCorruptedFiles checks that LoadJSON correctly handles various
// types of corruption that can occur during the saving process.
func TestLoadJSONCorruptedFiles(t *testing.T) {
	if testing.Short() {
		t.SkipNow()
	}
	// Define the test object that will be getting loaded.
	testMeta := Metadata{"Test Struct", "v1.2.1"}
	type testStruct struct {
		One   string
		Two   uint64
		Three []byte
	}
	obj1 := testStruct{"dog", 25, []byte("more dog")}
	var obj2 testStruct

	// Try loading a file with a bad checksum.
	err := LoadJSON(testMeta, &obj2, filepath.Join("testdata", "badchecksum.json"))
	if err == nil {
		t.Error("bad checksum should have failed")
	}
	// Try loading a file where only the main has a bad checksum.
	err = LoadJSON(testMeta, &obj2, filepath.Join("testdata", "badchecksummain.json"))
	if err != nil {
		t.Error("bad checksum main failed:", err)
	}
	// Verify equivalence.
	if obj2.One != obj1.One {
		t.Error("persist mismatch")
	}
	if obj2.Two != obj1.Two {
		t.Error("persist mismatch")
	}
	if !bytes.Equal(obj2.Three, obj1.Three) {
		t.Error("persist mismatch")
	}
	if obj2.One != "dog" {
		t.Error("persist mismatch")
	}
	if obj2.Two != 25 {
		t.Error("persist mismatch")
	}
	if !bytes.Equal(obj2.Three, []byte("more dog")) {
		t.Error("persist mismatch")
	}

	// Try loading a file with a manual checksum.
	err = LoadJSON(testMeta, &obj2, filepath.Join("testdata", "manual.json"))
	if err != nil {
		t.Error("bad checksum should have failed")
	}
	// Verify equivalence.
	if obj2.One != obj1.One {
		t.Error("persist mismatch")
	}
	if obj2.Two != obj1.Two {
		t.Error("persist mismatch")
	}
	if !bytes.Equal(obj2.Three, obj1.Three) {
		t.Error("persist mismatch")
	}
	if obj2.One != "dog" {
		t.Error("persist mismatch")
	}
	if obj2.Two != 25 {
		t.Error("persist mismatch")
	}
	if !bytes.Equal(obj2.Three, []byte("more dog")) {
		t.Error("persist mismatch")
	}

	// Try loading a corrupted main file.
	err = LoadJSON(testMeta, &obj2, filepath.Join("testdata", "corruptmain.json"))
	if err != nil {
		t.Error("couldn't load corrupted main:", err)
	}
	// Verify equivalence.
	if obj2.One != obj1.One {
		t.Error("persist mismatch")
	}
	if obj2.Two != obj1.Two {
		t.Error("persist mismatch")
	}
	if !bytes.Equal(obj2.Three, obj1.Three) {
		t.Error("persist mismatch")
	}
	if obj2.One != "dog" {
		t.Error("persist mismatch")
	}
	if obj2.Two != 25 {
		t.Error("persist mismatch")
	}
	if !bytes.Equal(obj2.Three, []byte("more dog")) {
		t.Error("persist mismatch")
	}

	// Try loading a corrupted temp file.
	err = LoadJSON(testMeta, &obj2, filepath.Join("testdata", "corrupttemp.json"))
	if err != nil {
		t.Error("couldn't load corrupted main:", err)
	}
	// Verify equivalence.
	if obj2.One != obj1.One {
		t.Error("persist mismatch")
	}
	if obj2.Two != obj1.Two {
		t.Error("persist mismatch")
	}
	if !bytes.Equal(obj2.Three, obj1.Three) {
		t.Error("persist mismatch")
	}
	if obj2.One != "dog" {
		t.Error("persist mismatch")
	}
	if obj2.Two != 25 {
		t.Error("persist mismatch")
	}
	if !bytes.Equal(obj2.Three, []byte("more dog")) {
		t.Error("persist mismatch")
	}

	// Try loading a file with no temp, and no checksum.
	err = LoadJSON(testMeta, &obj2, filepath.Join("testdata", "nochecksum.json"))
	if err != nil {
		t.Error("couldn't load no checksum:", err)
	}
	// Verify equivalence.
	if obj2.One != obj1.One {
		t.Error("persist mismatch")
	}
	if obj2.Two != obj1.Two {
		t.Error("persist mismatch")
	}
	if !bytes.Equal(obj2.Three, obj1.Three) {
		t.Error("persist mismatch")
	}
	if obj2.One != "dog" {
		t.Error("persist mismatch")
	}
	if obj2.Two != 25 {
		t.Error("persist mismatch")
	}
	if !bytes.Equal(obj2.Three, []byte("more dog")) {
		t.Error("persist mismatch")
	}
}
