package confluence

import models "github.com/ctreminiom/go-atlassian/v2/pkg/infra/models"

type ListSpacesRequest struct {
	Options *models.GetSpacesOptionSchemeV2
	Cursor  string
	Limit   int
}

type SpaceRequest struct {
	SpaceID           int
	DescriptionFormat string
}

type ListPagesRequest struct {
	Options *models.PageOptionsScheme
	Cursor  string
	Limit   int
}

type PageRequest struct {
	PageID  int
	Format  string
	Draft   bool
	Version int
}

type ListPagesBySpaceRequest struct {
	SpaceID int
	Cursor  string
	Limit   int
}

type ListChildPagesRequest struct {
	PageID int
	Cursor string
	Limit  int
}

type ListPagesByLabelRequest struct {
	LabelID int
	Sort    string
	Cursor  string
	Limit   int
}

type ListAttachmentsRequest struct {
	EntityID   int
	EntityType string
	Options    *models.AttachmentParamsScheme
	Cursor     string
	Limit      int
}

type AttachmentRequest struct {
	AttachmentID string
	VersionID    int
	SerializeIDs bool
}

type ListCustomContentRequest struct {
	Type    string
	Options *models.CustomContentOptionsScheme
	Cursor  string
	Limit   int
}

type CustomContentRequest struct {
	CustomContentID int
	Format          string
	VersionID       int
}

type Response = models.ResponseScheme
type Space = models.SpaceSchemeV2
type SpaceChunk = models.SpaceChunkV2Scheme
type Page = models.PageScheme
type PageChunk = models.PageChunkScheme
type ChildPageChunk = models.ChildPageChunkScheme
type Attachment = models.AttachmentScheme
type AttachmentPage = models.AttachmentPageScheme
type CustomContent = models.CustomContentScheme
type CustomContentPage = models.CustomContentPageScheme
