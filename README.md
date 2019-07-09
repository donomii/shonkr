# Shonkr


A slightly shonky text editor

# An experiment in text rendering

Shonkr is an experiment in adding interesting features to a text editor.  Unfortunately, I first had to write a text renderer.

The text renderer takes some text, and a width and height, and creates an image with the text laid out.  It can be used to add a text editor box to Gui frameworks that don't support it.

Now that the fundamentals are working, I plan to start experimenting with the interesting features like arbitrary text placement, multiple document views, and help annotations.  The goal is to turn the text into an active document that monitors and assists the user.


# Install

There is currently a test framework in the v3 directory that is capable of loading and editing a file.

    go get github.com/donomii/shonkr
    go build github.com/donomii/shonkr/v3


# Use

    ./v3

V3 will display a list of files in the current directory, you can click on them to edit.  Changes will be automatically saved when you open a new file.
