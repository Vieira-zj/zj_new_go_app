package model

const TableNameBlob = "blob"

type Blob struct {
	ID    int32  `gorm:"column:id;primaryKey;autoIncrement:true;type:int" json:"id"` // id
	Text  string `gorm:"column:string;type:text;default:null" json:"text"`           // text
	Bytes []byte `gorm:"column:bytes;type:blob" json:"bytes"`                        // bytes
}

func (*Blob) TableName() string {
	return TableNameBlob
}
