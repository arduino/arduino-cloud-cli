package folder

import (
	"time"

	"github.com/arduino/arduino-cli/table"
)

type PathNode struct {
	// ID of the folder that makes up this path node
	FolderId string `json:"folder_id"`
	// Name of the folder that makes up this path node
	FolderName string `json:"folder_name"`
}

type Folder struct {
	ID        string     `json:"id"`
	Name      string     `json:"name"`
	CreatedAt time.Time  `json:"created_at"`
	UpdatedAt time.Time  `json:"updated_at"`
	Parent    *PathNode  `json:"parent,omitempty"`
	Path      []PathNode `json:"path,omitempty"`
}

type Folders struct {
	Folders []Folder `json:"folders"`
}

func (r *Folders) Data() interface{} {
	return r
}

func (r *Folders) String() string {
	if len(r.Folders) == 0 {
		return ""
	}
	t := table.New()
	t.SetHeader("Folder ID", "Name", "Created At", "Updated At", "Parent", "Path")

	// Now print the table
	for _, fol := range r.Folders {
		parent := ""
		if fol.Parent != nil {
			parent = fol.Parent.FolderName
		}
		if fol.Path != nil {
			for _, p := range fol.Path {
				parent += "> " + p.FolderName
			}
		}
		path := ""
		line := []any{fol.ID, fol.Name, fol.CreatedAt.Format(time.RFC3339), fol.UpdatedAt.Format(time.RFC3339), parent, path}
		t.AddRow(line...)
	}

	return t.Render()
}
