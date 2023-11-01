package form

import (
	"context"
	"encoding/json"
	"io"

	"github.com/go-git/go-billy/v5"
	"github.com/gov4git/lib4git/must"
	"github.com/gov4git/lib4git/ns"
)

type Form interface{}

type Forms []Form

func ToForms[T Form](xs []T) Forms {
	r := make(Forms, len(xs))
	for i, x := range xs {
		r[i] = x
	}
	return r
}

type Map = map[string]Form

type None struct{}

func SprintJSON(form Form) string {
	data, err := json.MarshalIndent(form, "", "   ")
	if err != nil {
		panic(err)
	}
	return string(data)
}

func Encode[F Form](ctx context.Context, w io.Writer, f F) error {
	return json.NewEncoder(w).Encode(f)
}

func Decode[F Form](ctx context.Context, r io.Reader) (form F, err error) {
	err = json.NewDecoder(r).Decode(&form)
	return form, err
}

func DecodeInto(ctx context.Context, r io.Reader, into Form) error {
	return json.NewDecoder(r).Decode(into)
}

func EncodeBytes[F Form](ctx context.Context, form F) ([]byte, error) {
	return json.MarshalIndent(form, "", "   ")
}

func DecodeBytes[F Form](ctx context.Context, data []byte) (form F, err error) {
	err = json.Unmarshal(data, &form)
	return form, err
}

func DecodeBytesInto(ctx context.Context, data []byte, into Form) error {
	return json.Unmarshal(data, into)
}

func EncodeToFile[F Form](ctx context.Context, fs billy.Filesystem, path ns.NS, form F) error {
	file, err := fs.Create(path.GitPath())
	if err != nil {
		return err
	}
	defer file.Close()
	return Encode(ctx, file, form)
}

func DecodeFromFile[F Form](ctx context.Context, fs billy.Filesystem, path ns.NS) (form F, err error) {
	file, err := fs.Open(path.GitPath())
	if err != nil {
		return form, err
	}
	defer file.Close()
	return Decode[F](ctx, file)
}

func DecodeFromFileInto(ctx context.Context, fs billy.Filesystem, path ns.NS, into Form) error {
	file, err := fs.Open(path.GitPath())
	if err != nil {
		return err
	}
	defer file.Close()
	return DecodeInto(ctx, file, into)
}

func ToFile[F Form](ctx context.Context, fs billy.Filesystem, path ns.NS, form F) {
	if err := EncodeToFile(ctx, fs, path, form); err != nil {
		must.Panic(ctx, err)
	}
}

func FromFile[F Form](ctx context.Context, fs billy.Filesystem, path ns.NS) F {
	f, err := DecodeFromFile[F](ctx, fs, path)
	if err != nil {
		must.Panic(ctx, err)
	}
	return f
}

func FromFileInto(ctx context.Context, fs billy.Filesystem, path ns.NS, into Form) {
	must.NoError(ctx, DecodeFromFileInto(ctx, fs, path, into))
}
