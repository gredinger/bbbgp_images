package main

import (
	"bytes"
	"fmt"
	"image"
	"image/png"
	"io"
	"log"
	"math/rand" // wrong random lib
	"net/http"
	"os"
	"slices"
	"time"

	"github.com/disintegration/imaging"
	"github.com/fogleman/gg"
	"github.com/gorilla/mux"
)

var validPassword string

const (
	DayState = iota
	Normal
	Cancelled
	Pizza
)

func main() {
	var valid bool
	validPassword, valid = os.LookupEnv("password")
	if !valid {
		panic("Please set a password in env vars")
	}
	r := mux.NewRouter()
	r.HandleFunc("/meeting", DrawHandler(Normal))
	r.HandleFunc("/cancelled", DrawHandler(Cancelled))
	r.HandleFunc("/pizza", DrawHandler(Pizza))
	r.HandleFunc("/upload", UploadHandler)
	http.Handle("/", r)

	srv := &http.Server{
		Handler: r,
		Addr:    "0.0.0.0:8000",
		// Good practice: enforce timeouts for servers you create!
		WriteTimeout: 300 * time.Second,
		ReadTimeout:  60 * time.Second,
	}

	log.Fatal(srv.ListenAndServe())
}

func DrawHandler(ImageState int) func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		buff := new(bytes.Buffer)
		png.Encode(buff, drawImage(ImageState))
		w.Write(buff.Bytes())
	}
}

// upload
var inputForm = `<html>
<head><title>Picture Upload</title></head>
<body>
<form method="POST" enctype="multipart/form-data" action='/upload' >
<label for="password">Password:</label><input type="text" id="password" name="password" /><br />
<input type='file' id='myFile' name='myFile'><br />
<input type='hidden' id='hidden' name='hidden' value='set'>
<input type='submit'>
</form>
</body>
</html>`

func UploadHandler(w http.ResponseWriter, r *http.Request) {
	r.ParseForm()
	r.ParseMultipartForm(1024 * 1024 * 1024 * 5)
	if r.PostFormValue("hidden") != "set" {
		fmt.Fprint(w, inputForm)
		return
	}
	if r.PostFormValue("password") != validPassword {
		fmt.Fprint(w, "<a href='/upload'>Please try again with a valid password</a>")
		return
	}
	file, handler, err := r.FormFile("myFile")
	if err != nil {
		log.Fatal(err)
	}
	dst, err := os.Create("img/" + handler.Filename)
	if err != nil {
		log.Fatal(err)
	}
	defer dst.Close()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// Copy the uploaded file to the created file on the filesystem
	if _, err := io.Copy(dst, file); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	fmt.Fprint(w, "<html>File uploaded successfully. <a href='/upload'>Upload another</a></html>")

}

func drawImage(state int) image.Image {
	// First meeting time
	fm := time.Date(2024, 2, 1, 15, 30, 0, 0, time.FixedZone("Eastern", -5*3600)) // UTC server time
	// Meet counting + 1 = (today - fm) in weeks (hours/days/weeks)
	mc := int64((time.Since(fm).Hours()/24)/14) + 1
	// Take the meeting count + 1, convert it to weeks; add it to the first meeting date
	nm := fm.Add(time.Hour * time.Duration(mc) * (24 * 14))
	// Next meeting ending time is 2 hours after the next meeting start
	nme := fm.Add(time.Hour * 2)
	// time format for next meeting
	nmt := nm.Format("3:04")
	// time format for next meeting end
	nmet := nme.Format("3:04 pm")
	// time format for next meeting day
	nmd := nm.Format("Monday")
	td := fmt.Sprintf("%v - %v", nmt, nmet) // time and day
	fd := nm.Format("January _2, 2006")     // full date
	// get all the images that have been saved
	f, err := os.ReadDir("img")
	if err != nil {
		panic(err)
	}
	// page size, width height
	pw := float64(816)
	ph := float64(1056)

	//generate some randoms
	rnl := []int{}     //random number list
	for len(rnl) < 5 { // need 5 random pictures
		rn := rand.Intn(len(f))
		if !slices.Contains(rnl, rn) {
			rnl = append(rnl, rn)
		}
	}
	images := []image.Image{}
	for i, x := range rnl {
		img, err := gg.LoadImage("img/" + f[x].Name())
		if err != nil {
			panic(err)
		}
		if i+1 != len(rnl) { // last image is largest
			img = imaging.Resize(img, int(pw*.15), int(ph*.15), imaging.Lanczos)
		} else {
			img = imaging.Resize(img, int(pw*.3), int(ph*.3), imaging.Lanczos)
		}
		images = append(images, img)
	}

	dc := gg.NewContext(int(pw), int(ph))
	dc.SetRGB(1, 1, 1)
	dc.Clear()
	dc.SetRGB(0, 0, 0)
	// output
	if err := dc.LoadFontFace("arial.ttf", 48); err != nil {
		panic(err)
	}
	dc.DrawImage(images[0], int(pw*.03), int(ph*.15))
	dc.DrawImage(images[1], int(pw*.81), int(ph*.15))
	dc.DrawImage(images[2], int(pw*.03), int(ph*.6))
	dc.DrawImage(images[3], int(pw*.80), int(ph*.6))
	dc.DrawImage(images[4], int(pw*.35), int(ph*.68))
	dc.DrawStringAnchored("Board of Bored Board Game Players", pw*.5, ph*.1, .5, .5)
	switch state {
	case Pizza:
		dc.DrawStringAnchored("Free Pizza", pw*.17, ph*.85, .5, .5)
		dc.DrawStringAnchored("and Soda!", pw*.17, ph*.90, .5, .5)
		fallthrough
	case Normal:
		dc.DrawStringAnchored("Join us for a day of fun", pw*.5, ph*.25, .5, .5)
	case Cancelled:
		dc.DrawStringAnchored("Meeting will not be held", pw*.5, ph*.25, .5, .5)
	}
	dc.DrawStringAnchored(nmd, pw*.5, ph*.35, .5, .5)
	dc.DrawStringAnchored(td, pw*.5, ph*.4, .5, .5)
	dc.DrawStringAnchored(fd, pw*.5, ph*.45, .5, .5)
	dc.DrawStringAnchored("Harmon Meeting Room", pw*.5, ph*.55, .5, .5)
	dc.DrawStringAnchored("Local History Center", pw*.5, ph*.6, .5, .5)
	dc.DrawStringAnchored("Bryan, Ohio", pw*.5, ph*.65, .5, .5)

	if err := dc.LoadFontFace("arial.ttf", 24); err != nil {
		panic(err)
	}
	dc.DrawStringAnchored("More info:", pw*.9, ph*.9, .5, .5)
	dc.DrawStringAnchored("bbbgp.org", pw*.9, ph*.93, .5, .5)

	return dc.Image()

}
