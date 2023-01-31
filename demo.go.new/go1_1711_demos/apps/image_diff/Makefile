work_dir = /tmp/test

prepare:
	cp testdata/before.png $(work_dir)/before.png
	cp testdata/after.png $(work_dir)/after.png

debug:
	go run . -output=$(work_dir)/diff.png $(work_dir)/before.png $(work_dir)/after.png
