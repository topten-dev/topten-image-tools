package screens

import (
	"image/color"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/widget"
	"github.com/topten-dev/topten-image-tools/core"
)

// imageKind represents the user's answer to the first wizard question.
type imageKind int

const (
	kindUnset        imageKind = iota
	kindPhoto                  // Photos, natural imagery → JPG
	kindGraphic                // Text, logos, sharp edges → PNG
	kindTransparent            // Has transparent areas → PNG
	kindBanner                 // Hero/banner, needs follow-up
)

// Wizard renders a conversational screen that guides the user towards the right
// output format. anyAlpha is true when at least one source file already has an
// alpha channel.
func Wizard(
	w fyne.Window,
	anyAlpha bool,
	onConfirm func(core.Format),
	onBack func(),
) fyne.CanvasObject {
	var chosenKind = kindUnset
	var bannerHasText *bool // nil = not yet answered
	var chosenFormat core.Format

	// ── Recommendation card (shown after all Q&A is complete) ──────────────
	recBox := container.NewVBox()

	// ── State machine ─────────────────────────────────────────────────────
	var renderRecommendation func()
	var renderBannerQ func()
	var content *fyne.Container

	confirmBtn := widget.NewButton("Use this format →", nil)
	confirmBtn.Importance = widget.HighImportance
	confirmBtn.Disable()

	renderRecommendation = func() {
		recBox.Objects = nil

		var fmt core.Format
		var reason, emoji string

		switch chosenKind {
		case kindPhoto:
			fmt = core.FormatJPEG
			reason = "Photos and natural images compress efficiently as JPG without noticeable quality loss."
			emoji = "📷"
		case kindGraphic:
			fmt = core.FormatPNG
			reason = "Images with text, logos, or sharp lines preserve quality best as PNG (lossless)."
			emoji = "✏️"
		case kindTransparent:
			fmt = core.FormatPNG
			reason = "PNG is the only web-safe format that supports transparency."
			emoji = "🔍"
		case kindBanner:
			if bannerHasText != nil && *bannerHasText {
				fmt = core.FormatPNG
				reason = "Banners with text or logos should be saved as PNG to keep lettering crisp."
				emoji = "📢"
			} else if bannerHasText != nil && !*bannerHasText {
				fmt = core.FormatJPEG
				reason = "Purely photographic banners compress well as JPG, keeping file sizes small."
				emoji = "🌅"
			} else {
				return // still waiting for banner follow-up
			}
		default:
			return
		}

		chosenFormat = fmt

		// Override: detected alpha forces PNG.
		if anyAlpha && fmt == core.FormatJPEG {
			fmt = core.FormatPNG
			reason += " (Override: we detected that at least one source file has a transparent background.)"
			chosenFormat = fmt
		}

		badge := canvas.NewText(string(fmt)+" recommended "+emoji, color.NRGBA{R: 255, G: 255, B: 255, A: 255})
		badge.TextSize = 16
		badge.TextStyle = fyne.TextStyle{Bold: true}

		reasonLabel := widget.NewLabel(reason)
		reasonLabel.Wrapping = fyne.TextWrapWord

		overridePNG := widget.NewButton("Use PNG instead", func() {
			chosenFormat = core.FormatPNG
			confirmBtn.Enable()
		})
		overrideJPG := widget.NewButton("Use JPG instead", func() {
			chosenFormat = core.FormatJPEG
			confirmBtn.Enable()
		})
		if fmt == core.FormatPNG {
			overridePNG.Disable()
		} else {
			overrideJPG.Disable()
		}

		bg := canvas.NewRectangle(color.NRGBA{R: 0, G: 100, B: 90, A: 220})
		bg.CornerRadius = 8

		recCard := container.NewStack(
			bg,
			container.NewPadded(container.NewVBox(
				container.NewCenter(badge),
				reasonLabel,
				container.NewCenter(container.NewHBox(overridePNG, overrideJPG)),
			)),
		)

		recBox.Add(recCard)
		recBox.Refresh()
		confirmBtn.Enable()
	}

	renderBannerQ = func() {
		content.Objects = []fyne.CanvasObject{
			bannerFollowUp(func(hasText bool) {
				bannerHasText = &hasText
				content.Objects = []fyne.CanvasObject{mainQuestion(func(k imageKind) {
					chosenKind = k
					if k == kindBanner {
						renderBannerQ()
					} else {
						renderRecommendation()
					}
				})}
				content.Add(recBox)
				renderRecommendation()
				content.Refresh()
			}),
			recBox,
		}
		content.Refresh()
	}

	// ── Main question ──────────────────────────────────────────────────────
	mainQ := mainQuestion(func(k imageKind) {
		chosenKind = k
		if k == kindBanner {
			renderBannerQ()
		} else {
			renderRecommendation()
		}
	})

	content = container.NewVBox(mainQ, recBox)

	confirmBtn.OnTapped = func() {
		onConfirm(chosenFormat)
	}

	// Pre-select transparent when detected.
	if anyAlpha {
		chosenKind = kindTransparent
		renderRecommendation()
	}

	title := widget.NewLabelWithStyle("What best describes your images?", fyne.TextAlignCenter, fyne.TextStyle{Bold: true})
	backBtn := widget.NewButton("← Back", onBack)

	return container.NewPadded(container.NewBorder(
		container.NewVBox(title),
		container.NewCenter(container.NewHBox(backBtn, confirmBtn)),
		nil, nil,
		container.NewVScroll(content),
	))
}

// mainQuestion renders the four image-kind option buttons.
func mainQuestion(onPick func(imageKind)) fyne.CanvasObject {
	opts := []struct {
		label string
		desc  string
		kind  imageKind
	}{
		{"📷  Photos or natural images", "Landscapes, portraits, product shots, gradients.", kindPhoto},
		{"✏️  Graphics with text or logos", "Screenshots, infographics, logos, or anything with sharp lines.", kindGraphic},
		{"🔍  Images with transparent background", "PNGs with see-through areas (e.g. icons, cutouts).", kindTransparent},
		{"📢  Website hero banners / featured images", "Large prominent images used at the top of pages or articles.", kindBanner},
	}

	box := container.NewVBox()
	for _, o := range opts {
		o := o // capture
		lbl := widget.NewLabelWithStyle(o.label, fyne.TextAlignLeading, fyne.TextStyle{Bold: true})
		desc := widget.NewLabel(o.desc)
		desc.TextStyle = fyne.TextStyle{Italic: true}

		btn := widget.NewButton("Select", func() { onPick(o.kind) })
		btn.Importance = widget.LowImportance

		bg := canvas.NewRectangle(color.NRGBA{R: 35, G: 38, B: 46, A: 255})
		bg.CornerRadius = 6

		row := container.NewStack(
			bg,
			container.NewPadded(container.NewBorder(nil, nil, nil, btn,
				container.NewVBox(lbl, desc),
			)),
		)
		box.Add(row)
	}
	return box
}

// bannerFollowUp asks the single follow-up question for hero banner images.
func bannerFollowUp(onAnswer func(hasText bool)) fyne.CanvasObject {
	label := widget.NewLabelWithStyle("Do your banners contain text overlays or logos?", fyne.TextAlignLeading, fyne.TextStyle{Bold: true})

	yesBtn := widget.NewButton("Yes — they have text or logos", func() { onAnswer(true) })
	yesBtn.Importance = widget.MediumImportance
	noBtn := widget.NewButton("No — purely photographic", func() { onAnswer(false) })
	noBtn.Importance = widget.MediumImportance

	bg := canvas.NewRectangle(color.NRGBA{R: 50, G: 44, B: 35, A: 255})
	bg.CornerRadius = 6

	return container.NewStack(
		bg,
		container.NewPadded(container.NewVBox(label, container.NewHBox(yesBtn, noBtn))),
	)
}
