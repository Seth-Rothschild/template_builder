local images = dofile("filters/images.lua")

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

assert_true(images.image_exists("testdata/image.png"), "image_exists should be true for a file that exists")
assert_true(not images.image_exists("testdata/does-not-exist.png"), "image_exists should be false for a missing file")

assert_equal("testdata/image.png", images.resolve_path("testdata/image.png"),
  "resolve_path should leave a relative path unchanged when IMAGE_DIR is unset")
assert_equal("/abs/image.png", images.resolve_path("/abs/image.png"),
  "resolve_path should leave an already-absolute path unchanged")

local original_input_files = PANDOC_STATE.input_files
PANDOC_STATE.input_files = { "testdata/e2e.md" }

assert_equal("testdata", images.source_dir(), "source_dir should return the markdown file's directory")
assert_equal("testdata/image.png", images.check_path("image.png"),
  "check_path should resolve a bare src against the markdown file's directory")
assert_equal("image.png", images.resolve_path("image.png"),
  "resolve_path should leave the src unchanged for embedding, since tectonic looks next to the .tex file")

local present_bare_figure_src = "image.png"
local present_bare_figure = pandoc.Figure(
  { pandoc.Plain({ pandoc.Image({ pandoc.Str("a bare src image") }, present_bare_figure_src) }) },
  pandoc.Caption(pandoc.Blocks({ pandoc.Plain({ pandoc.Str("a bare src image") }) }))
)
local expected_bare = "\\begin{figure}[H]\n"
    .. "\\centering\n"
    .. "\\includegraphics[width=\\linewidth,height=\\textheight,keepaspectratio]{image.png}\n"
    .. "\\caption{a bare src image}\n"
    .. "\\end{figure}"
assert_equal(expected_bare, images.build_figure(present_bare_figure),
  "build_figure should find an image whose src is relative to the markdown file, but embed the original bare src")

PANDOC_STATE.input_files = original_input_files

local function figure(src, caption_text)
  local caption_inlines = { pandoc.Str(caption_text) }
  local image = pandoc.Image(caption_inlines, src)
  local caption = pandoc.Caption(pandoc.Blocks({ pandoc.Plain(caption_inlines) }))
  return pandoc.Figure({ pandoc.Plain({ image }) }, caption)
end

local present_figure = figure("testdata/image.png", "a test image")
local expected_present = "\\begin{figure}[H]\n"
    .. "\\centering\n"
    .. "\\includegraphics[width=\\linewidth,height=\\textheight,keepaspectratio]{testdata/image.png}\n"
    .. "\\caption{a test image}\n"
    .. "\\end{figure}"
assert_equal(expected_present, images.build_figure(present_figure), "build_figure with an existing image")

local missing_figure = figure("testdata/does-not-exist.png", "a missing image")
local expected_missing = "\\begin{figure}[H]\n"
    .. "\\centering\n"
    .. "\\refstepcounter{figure}Figure \\thefigure{} (missing image): a missing image\n"
    .. "\\end{figure}"
assert_equal(expected_missing, images.build_figure(missing_figure), "build_figure with a missing image")

print("all tests passed")
