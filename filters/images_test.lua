local images = dofile("filters/images.lua")[1]

local function assert_equal(expected, actual, message)
  if expected ~= actual then
    error(string.format("%s: expected %q, got %q", message, expected, actual))
  end
end

local function assert_true(condition, message)
  if not condition then
    error(message)
  end
end

local function figure(src, caption_text)
  local caption_inlines = { pandoc.Str(caption_text) }
  local image = pandoc.Image(caption_inlines, src)
  local caption = pandoc.Caption(pandoc.Blocks({ pandoc.Plain(caption_inlines) }))
  return pandoc.Figure({ pandoc.Plain({ image }) }, caption)
end

assert_true(images.image_exists("testdata/image.png"), "image_exists should be true for a file that exists")
assert_true(not images.image_exists("testdata/does-not-exist.png"), "image_exists should be false for a missing file")

local original_input_files = PANDOC_STATE.input_files
PANDOC_STATE.input_files = { "testdata/e2e.md" }

assert_equal("testdata", images.source_dir(), "source_dir should return the markdown file's directory")
assert_equal("testdata/image.png", images.disk_path("image.png"),
  "disk_path should resolve a relative src against the markdown file's directory")
assert_equal("/abs/image.png", images.disk_path("/abs/image.png"),
  "disk_path should leave an absolute src unchanged")

local expected_present = "\\begin{figure}[H]\n"
    .. "\\centering\n"
    .. "\\includegraphics[width=\\linewidth,height=\\textheight,keepaspectratio]{image.png}\n"
    .. "\\caption{a test image}\n"
    .. "\\end{figure}"
assert_equal(expected_present, images.build_figure(figure("image.png", "a test image")),
  "build_figure should find an image next to the markdown file and embed its original src")

local expected_missing = "\\begin{figure}[H]\n"
    .. "\\centering\n"
    .. "\\refstepcounter{figure}Figure \\thefigure{} (missing image): a missing image\n"
    .. "\\end{figure}"
assert_equal(expected_missing, images.build_figure(figure("does-not-exist.png", "a missing image")),
  "build_figure should mark a missing image")

PANDOC_STATE.input_files = original_input_files

print("all tests passed")
