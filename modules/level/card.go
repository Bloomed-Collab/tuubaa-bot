package level

import (
	"bytes"
	"fmt"
	"image"
	"image/color"
	"image/png"
	"math"
	"net/http"
	"os"
	"time"

	"github.com/fogleman/gg"
	"github.com/golang/freetype/truetype"
	xdraw "golang.org/x/image/draw"
	"golang.org/x/image/font/gofont/goregular"
)

const backgroundImageURL = "https://i.postimg.cc/7YK0SfsQ/image-2.png"
const leaderboardBackgroundLocalPath = "assets/level/awesome.png"

type leaderboardEntry struct {
	Rank        int
	UserID      string
	XP          int64
	DisplayName string
	AvatarURL   string
}

func buildLeaderboardImage(entries []leaderboardEntry) (*bytes.Buffer, error) {
	const (
		W        = 680
		rowH     = 68
		avatarD  = 52
		padX     = 14
		padY     = 14
		rankSize = 36.0
	)

	H := padY + 10*rowH + padY

	type imgResult struct {
		idx int // 0 = background, 1..N = avatar for entry[idx-1]
		img image.Image
	}
	total := 1 + len(entries)
	ch := make(chan imgResult, total)

	fetch := func(idx int, url string, timeout time.Duration) {
		if url == "" {
			ch <- imgResult{idx, nil}
			return
		}
		client := &http.Client{Timeout: timeout}
		req, _ := http.NewRequest(http.MethodGet, url, nil)
		req.Header.Set("User-Agent", "tuubaa-bot/level-card")
		resp, err := client.Do(req)
		if err != nil {
			ch <- imgResult{idx, nil}
			return
		}
		defer resp.Body.Close()
		img, _, err := image.Decode(resp.Body)
		if err != nil {
			ch <- imgResult{idx, nil}
			return
		}
		ch <- imgResult{idx, img}
	}

	var bgLocal image.Image
	if data, err := os.ReadFile(leaderboardBackgroundLocalPath); err == nil {
		if img, _, decErr := image.Decode(bytes.NewReader(data)); decErr == nil {
			bgLocal = img
		}
	}
	if bgLocal != nil {
		ch <- imgResult{0, bgLocal}
	} else {
		go fetch(0, backgroundImageURL, 8*time.Second)
	}
	for i, e := range entries {
		go fetch(i+1, e.AvatarURL, 1500*time.Millisecond)
	}

	fetched := make([]image.Image, total)
	for i := 0; i < total; i++ {
		r := <-ch
		fetched[r.idx] = r.img
	}
	bgImg := fetched[0]
	avatarImgs := fetched[1:]

	dc := gg.NewContext(W, H)

	if bgImg != nil {
		scaled := image.NewRGBA(image.Rect(0, 0, W, H))
		xdraw.BiLinear.Scale(scaled, scaled.Bounds(), bgImg, bgImg.Bounds(), xdraw.Over, nil)
		dc.DrawImage(scaled, 0, 0)
	} else {
		dc.SetRGB(0.07, 0.08, 0.11)
		dc.Clear()
	}

	dc.SetRGBA(0.04, 0.04, 0.07, 0.58)
	dc.DrawRectangle(0, 0, float64(W), float64(H))
	dc.Fill()

	for idx, e := range entries {
		rowTop := float64(padY + idx*rowH)
		centerY := rowTop + float64(rowH)/2

		avCX := float64(padX + avatarD/2)
		avCY := centerY
		avImg := avatarImgs[idx]
		drawCircleAvatar(dc, avImg, avCX, avCY, avatarD)

		textX := float64(padX+avatarD) + 12
		cardFont(dc, 18)
		dc.SetRGB(1, 1, 1)
		dc.DrawString(e.DisplayName, textX, centerY-5)

		cardFont(dc, 13)
		dc.SetRGBA(0.78, 0.78, 0.78, 1.0)
		dc.DrawString(fmt.Sprintf("Level %d", calcLevel(e.XP)), textX, centerY+13)

		var rr, rg, rb float64
		switch e.Rank {
		case 1:
			rr, rg, rb = 1.0, 0.84, 0.0   // gold
		case 2:
			rr, rg, rb = 0.75, 0.75, 0.75 // silver
		case 3:
			rr, rg, rb = 0.80, 0.50, 0.20 // bronze
		default:
			rr, rg, rb = 0.85, 0.85, 0.85
		}
		rankText := fmt.Sprintf("#%d", e.Rank)
		cardFont(dc, rankSize)
		dc.SetRGB(rr, rg, rb)
		rw, _ := dc.MeasureString(rankText)
		dc.DrawString(rankText, float64(W)-rw-float64(padX), centerY+13)

		if idx < len(entries)-1 {
			dc.SetRGBA(1, 1, 1, 0.10)
			sepY := rowTop + float64(rowH)
			dc.DrawLine(textX, sepY, float64(W-padX), sepY)
			dc.SetLineWidth(0.5)
			dc.Stroke()
		}
	}

	buf := &bytes.Buffer{}
	if err := png.Encode(buf, dc.Image()); err != nil {
		return nil, err
	}
	return buf, nil
}

func drawCircleAvatar(dc *gg.Context, src image.Image, cx, cy float64, diameter int) {
	r := float64(diameter) / 2

	if src != nil {
		dst := image.NewRGBA(image.Rect(0, 0, diameter, diameter))
		xdraw.BiLinear.Scale(dst, dst.Bounds(), src, src.Bounds(), xdraw.Over, nil)

		clip := gg.NewContext(diameter, diameter)
		clip.DrawCircle(r, r, r)
		clip.Clip()
		clip.DrawImage(dst, 0, 0)

		dc.DrawImage(clip.Image(), int(cx-r), int(cy-r))
	} else {
		dc.SetRGBA(0.25, 0.25, 0.30, 1.0)
		dc.DrawCircle(cx, cy, r)
		dc.Fill()
	}

	dc.SetRGBA(1, 1, 1, 0.25)
	dc.DrawCircle(cx, cy, r+1)
	dc.SetLineWidth(1.5)
	dc.Stroke()
}

func buildRankCard(displayName, avatarURL string, level, rank int, xpCurrent, xpTotal int64) (*bytes.Buffer, error) {
	const (
		W      = 1024
		H      = 340
		cardR  = 34.0
		avSize = 220
		avX    = 42
	)
	avY := (H - avSize) / 2

	dc := gg.NewContext(W, H)

	dc.SetRGBA(0, 0, 0, 0)
	dc.Clear()

	dc.SetRGB(0.02, 0.03, 0.05)
	dc.DrawRoundedRectangle(0, 0, W, H, cardR)
	dc.Fill()

	dc.SetRGBA(0.08, 0.09, 0.12, 0.35)
	dc.DrawRoundedRectangle(10, 10, W-20, H-20, cardR-8)
	dc.Fill()

	avCX := float64(avX + avSize/2)
	avCY := float64(avY + avSize/2)
	avR := float64(avSize) / 2
	accentR, accentG, accentB := 0.34, 0.80, 0.61
	if avatar := fetchAvatarImage(avatarURL, avSize); avatar != nil {
		accentR, accentG, accentB = dominantAvatarAccent(avatar)
		clip := gg.NewContext(avSize, avSize)
		clip.DrawCircle(avR, avR, avR)
		clip.Clip()
		clip.DrawImage(avatar, 0, 0)
		dc.DrawImage(clip.Image(), avX, avY)
	} else {
		dc.SetRGBA(0.24, 0.25, 0.29, 1.0)
		dc.DrawCircle(avCX, avCY, avR)
		dc.Fill()
	}
	dc.SetRGBA(accentR, accentG, accentB, 0.85)
	dc.SetLineWidth(6)
	dc.DrawCircle(avCX, avCY, avR+1)
	dc.Stroke()

	const dotR = 16.0
	dotX := avCX + avR*0.62
	dotY := avCY + avR*0.62
	dc.SetRGB(0.22, 0.23, 0.28)
	dc.DrawCircle(dotX, dotY, dotR+6)
	dc.Fill()
	dc.SetRGB(accentR, accentG, accentB)
	dc.DrawCircle(dotX, dotY, dotR)
	dc.Fill()

	textX := float64(avX+avSize) + 36
	midY := float64(H) / 2

	cardFont(dc, 72)
	dc.SetRGB(0.86, 0.87, 0.89)
	dc.DrawString(displayName, textX, midY-6)

	cardFont(dc, 56)
	dc.SetRGB(accentR, accentG, accentB)
	dc.DrawString(fmt.Sprintf("Level %d", level), textX, midY+66)

	cardFont(dc, 54)
	dc.SetRGB(accentR, accentG, accentB)
	dc.DrawString(fmt.Sprintf("Exp:  %s / %s", fmtXP(xpCurrent), fmtXP(xpTotal)), textX, midY+126)

	const (
		ringCXOffset = 145.0
		outerR       = 92.0
		ringW        = 16.0
	)
	ringCX := float64(W) - ringCXOffset
	ringCY := midY

	progress := 0.0
	if xpTotal > 0 {
		progress = float64(xpCurrent) / float64(xpTotal)
		if progress > 1 {
			progress = 1
		}
	}

	dc.SetRGBA(accentR*0.35, accentG*0.35, accentB*0.35, 1.0)
	dc.SetLineWidth(ringW)
	dc.DrawArc(ringCX, ringCY, outerR-ringW/2, 0, 2*math.Pi)
	dc.Stroke()

	if progress > 0 {
		dc.SetRGBA(accentR, accentG, accentB, 1.0)
		dc.SetLineWidth(ringW)
		dc.DrawArc(ringCX, ringCY, outerR-ringW/2, -math.Pi/2, -math.Pi/2+progress*2*math.Pi)
		dc.Stroke()
	}

	rankText := fmt.Sprintf("#%d", rank)
	cardFont(dc, 56)
	dc.SetRGB(0.36, 0.37, 0.41)
	rw, rh := dc.MeasureString(rankText)
	dc.DrawString(rankText, ringCX-rw/2, ringCY+rh/2-4)

	buf := &bytes.Buffer{}
	if err := png.Encode(buf, dc.Image()); err != nil {
		return nil, err
	}
	return buf, nil
}

func fetchAvatarImage(url string, size int) image.Image {
	if url == "" {
		return nil
	}
	client := &http.Client{Timeout: 2 * time.Second}
	resp, err := client.Get(url)
	if err != nil {
		return nil
	}
	defer resp.Body.Close()

	src, _, err := image.Decode(resp.Body)
	if err != nil {
		return nil
	}

	dst := image.NewRGBA(image.Rect(0, 0, size, size))
	xdraw.BiLinear.Scale(dst, dst.Bounds(), src, src.Bounds(), xdraw.Over, nil)
	return dst
}

func drawRoundedAvatar(dc *gg.Context, url string, x, y float64, size int, radius float64) {
	fSize := float64(size)

	if url != "" {
		client := &http.Client{Timeout: 2 * time.Second}
		resp, err := client.Get(url)
		if err == nil {
			defer resp.Body.Close()
			src, _, err := image.Decode(resp.Body)
			if err == nil {
				dst := image.NewRGBA(image.Rect(0, 0, size, size))
				xdraw.BiLinear.Scale(dst, dst.Bounds(), src, src.Bounds(), xdraw.Over, nil)

				clip := gg.NewContext(size, size)
				clip.DrawRoundedRectangle(0, 0, fSize, fSize, radius)
				clip.Clip()
				clip.DrawImage(dst, 0, 0)

				dc.DrawImage(clip.Image(), int(x), int(y))
				return
			}
		}
	}
	dc.SetRGBA(0.25, 0.25, 0.30, 1.0)
	dc.DrawRoundedRectangle(x, y, fSize, fSize, radius)
	dc.Fill()
}

func fmtXP(raw int64) string {
	return fmt.Sprintf("%d", raw)
}

func dominantAvatarAccent(img image.Image) (float64, float64, float64) {
	if img == nil {
		return 0.34, 0.80, 0.61
	}

	b := img.Bounds()
	if b.Dx() == 0 || b.Dy() == 0 {
		return 0.34, 0.80, 0.61
	}

	type binKey uint16 // 5 bits per channel -> 15 bits used
	counts := make(map[binKey]int, 1024)

	stepX := maxInt(1, b.Dx()/48)
	stepY := maxInt(1, b.Dy()/48)

	for y := b.Min.Y; y < b.Max.Y; y += stepY {
		for x := b.Min.X; x < b.Max.X; x += stepX {
			c := color.NRGBAModel.Convert(img.At(x, y)).(color.NRGBA)
			if c.A < 180 {
				continue
			}

			if (c.R < 18 && c.G < 18 && c.B < 18) || (c.R > 240 && c.G > 240 && c.B > 240) {
				continue
			}

			r5 := uint16(c.R >> 3)
			g5 := uint16(c.G >> 3)
			b5 := uint16(c.B >> 3)
			k := binKey((r5 << 10) | (g5 << 5) | b5)
			counts[k]++
		}
	}

	if len(counts) == 0 {
		var sr, sg, sb, n float64
		for y := b.Min.Y; y < b.Max.Y; y += stepY {
			for x := b.Min.X; x < b.Max.X; x += stepX {
				c := color.NRGBAModel.Convert(img.At(x, y)).(color.NRGBA)
				if c.A < 180 {
					continue
				}
				sr += float64(c.R)
				sg += float64(c.G)
				sb += float64(c.B)
				n++
			}
		}
		if n <= 0 {
			return 0.34, 0.80, 0.61
		}
		return clamp01((sr / n) / 255), clamp01((sg / n) / 255), clamp01((sb / n) / 255)
	}

	var bestK binKey
	bestCount := -1
	for k, c := range counts {
		if c > bestCount {
			bestCount = c
			bestK = k
		}
	}

	r5 := (uint16(bestK) >> 10) & 31
	g5 := (uint16(bestK) >> 5) & 31
	b5 := uint16(bestK) & 31

	r := float64(r5*8+4) / 255.0
	g := float64(g5*8+4) / 255.0
	bl := float64(b5*8+4) / 255.0
	return clamp01(r), clamp01(g), clamp01(bl)
}

func clamp01(v float64) float64 {
	if v < 0 {
		return 0
	}
	if v > 1 {
		return 1
	}
	return v
}

func maxInt(a, b int) int {
	if a > b {
		return a
	}
	return b
}

func cardFont(dc *gg.Context, size float64) {
	data, err := os.ReadFile("assets/fonts/JetBrainsMono-Bold.ttf")
	if err == nil {
		f, err2 := truetype.Parse(data)
		if err2 == nil {
			dc.SetFontFace(truetype.NewFace(f, &truetype.Options{Size: size}))
			return
		}
	}
	f, _ := truetype.Parse(goregular.TTF)
	dc.SetFontFace(truetype.NewFace(f, &truetype.Options{Size: size}))
}
