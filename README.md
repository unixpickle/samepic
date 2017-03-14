# samepic

Suppose you have two images and you want to know if they're the "same" image. One image might be cropped, scaled, saturated, etc., but a human could still quickly tell they're the same image. How could you get a machine to do it?

# Results

Here are a bunch of algorithms and their performance on 100 test samples. It measures the algorithm's success rate on both positive samples and negative samples.

| Algorithm       | Positive success rate | Negative success rate |
| --------------- | --------------------- | --------------------- |
| Average Hashing | 28%                   | 100%                  |
| Color Profiling | 94%                   | 100%                  |
| Neural Networks | 92%                   | 100%                  |
| SquashComp      | 70%                   | 100%                  |

The algorithms were tuned to prevent false positives (i.e. errors when dealing with negative samples). This may or may not be desired, depending on use case.
