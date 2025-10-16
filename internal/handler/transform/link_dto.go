package transform

import (
    "service-short-link/internal/domain"
)

// LinkCreatedDTO represents minimal payload after creating link
type LinkCreatedDTO struct {
    ShortURL    string      `json:"short_url" example:"http://short.vieclam24h.vn/aB3x9K2"`
    QRCodeURL   string      `json:"qr_url" example:"http://short.vieclam24h.vn/qr/aB3x9K2"`
}


// ToLinkCreatedDTO builds the payload for create link response
func ToLinkCreatedDTO(resp *domain.CreateLinkResponse) LinkCreatedDTO {
    return LinkCreatedDTO{
        ShortURL:    resp.ShortURL,
        QRCodeURL:   resp.QRCodeURL,
    }
}
