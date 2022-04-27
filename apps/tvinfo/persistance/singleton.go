package persistance

var writer Writer

const persistanceFileName = "known_img_hashes.txt"

func init() {
	writer = NewFileWriter("known_img_hashes.txt")
}

func Append(str string) error {
	return writer.Append(str)
}

func Contains(str string) (bool, error) {
	return writer.Contains(str)
}
