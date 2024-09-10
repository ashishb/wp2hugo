package hugogenerator

import "path/filepath"

const _customCSS = `
.gallery {
	display: flex;
	flex-wrap: wrap;
}
.gallery figure,
.gallery figure img {
	text-align: center;
}
.gallery figure img {
	margin: 1rem auto;
}
.gallery-cols-1 figure {
	width: 100%;
}
.gallery-cols-2 figure {
	width: 50%;
}
.gallery-cols-3 figure {
	width: 33.3333333333%;
}
.gallery-cols-4 figure {
	width: 25%;
}
.gallery-cols-5 figure {
	width: 25%;
}
.gallery-cols-6 figure {
	width: 16.666666666%;
}
`

func setupCSS(siteDir string) error {
	err := appendFile(filepath.Join(siteDir, _outputCssFile), _customCSS)
	return err
}
