package samepic

// Rate computes the positive and negative success rates
// of a Samer on a sample set.
//
// Half of the time, the Samer will be given different
// images, and half of the time it will be given a pair
// of images generated by manipulating one image.
// In total, n tests (rounded down so n is even) will
// be run on the Samer.
//
// The result includes two ratios: the success rate on
// positive samples (i.e. same images), and the success
// rate on negative samples (i.e. different images).
//
// If samples returns an error at any point, Rate fails
// with the same error.
func Rate(samer Samer, samples Samples, manip Manipulator, n int) (posRate,
	negRate float64, err error) {
	var posCorrect int
	for i := 0; i < n/2; i++ {
		sample, err := samples.Random()
		if err != nil {
			return 0, 0, err
		}
		img1 := manip.Manipulate(sample)
		img2 := manip.Manipulate(sample)
		if samer.Same(img1, img2) {
			posCorrect++
		}
	}

	var negCorrect int
	for i := 0; i < n/2; i++ {
		img1, img2, err := samples.RandomPair()
		if err != nil {
			return 0, 0, err
		}
		if !samer.Same(img1, img2) {
			negCorrect++
		}
	}

	halfTotal := float64(n / 2)
	return float64(posCorrect) / halfTotal, float64(negCorrect) / halfTotal, nil
}
