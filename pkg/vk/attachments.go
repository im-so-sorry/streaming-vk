package vk

type AttachmentPhotoType struct {
	//https://vk.com/dev/objects/photo
	Id      int
	AlbumId int
	OwnerId int
	UserId  int
	Text    string
	Date    int
	Sizes   []struct {
		Type   string
		Url    string
		Width  int
		Height int
	}
	Width  int
	Height int
}

type AttachmentVideoType struct {
	//https://vk.com/dev/objects/video
	Id          int
	OwnerId     int
	Title       string
	Description string
	Duration    int // длительность видео в секундах

	Date int
	AddingDate int
}
