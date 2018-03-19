package pusher

// PhotoReader allows clients to read photos off of a connected camera.
// It only supports one connected camera at a time.
type PhotoReader struct {
	camera camera
}

// NewPhotoReader creates and returns a PhotoReader.
// Clients should call Close() when they are done with the PhotoReader.
func NewPhotoReader() *PhotoReader {
	pr := &PhotoReader{
		camera: &gphotoCamera{},
	}
	return pr
}

// GetFilenames returns a slice of filenames of all photos on the camera.
func (pr *PhotoReader) GetFilenames() ([]string, error) {
	filenames, err := pr.camera.listFilenames()
	return filenames, err
}

// ReadFile returns a byte slice containing the contents of a specific file on the camera.
func (pr *PhotoReader) ReadFile(filename string) ([]byte, error) {
	file, err := pr.camera.getFile(filename)
	return file, err
}

// Connected indicates whether there is currently a camera connected.
func (pr *PhotoReader) Connected() bool {
	return false
}

// CameraInfo returns some info about a connected camera.
func (pr *PhotoReader) CameraInfo() (map[string]string, error) {
	return map[string]string{}, nil
}

func (pr *PhotoReader) Close() {

}
