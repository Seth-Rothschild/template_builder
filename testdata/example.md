---
template_version: slick
title: Example document
subtitle: How Markdown Becomes a Typeset PDF
date: 2026-09-21
header-right: Design Note
footer-left: Document Pipeline Design
---

# Overview {#sec:overview}

*This document was written by an LLM as an example of what the builder can
produce. It exists to exercise diagrams, tables, and cross references in one
place, so read it as a sample of the output and not as a source of fact. The
pipeline it describes is real and matches the code. Every number in it is
invented: nothing in Section \ref{sec:measurements} was measured, and the
timings there are illustrative only.*

This note describes the pipeline that turns a Markdown file into a typeset PDF.
The stages are sketched in Figure \ref{fig:pipeline} on page \pageref{fig:pipeline},
the failure handling in Section \ref{sec:failure}, and the cost of each stage in
Section \ref{sec:measurements}.

Every diagram here is written as TikZ source inside the Markdown file. Nothing is
imported, and nothing was drawn by hand.

\begin{figure}[H]
\centering
\begin{tikzpicture}[
  node distance = 7mm,
  stage/.style = {draw=accent, thick, rounded corners=2pt, fill=accent!5,
                  minimum width=18mm, minimum height=9mm, align=center,
                  font=\small\bfseries, text=accent},
  artifact/.style = {draw=textgray!40, thick, dashed, rounded corners=2pt,
                     minimum width=18mm, minimum height=9mm, align=center,
                     font=\small, text=textgray},
  steering/.style = {artifact, font=\scriptsize, minimum width=24mm},
  flow/.style = {-{Stealth[length=2mm]}, draw=accent, thick}
]
\node[artifact] (md) {\texttt{input.md}};
\node[stage, right=of md] (pandoc) {pandoc};
\node[artifact, right=of pandoc] (tex) {\texttt{input.tex}};
\node[stage, right=of tex] (tectonic) {tectonic};
\node[artifact, right=of tectonic] (pdf) {\texttt{input.pdf}};

\draw[flow] (md) -- (pandoc);
\draw[flow] (pandoc) -- (tex);
\draw[flow] (tex) -- (tectonic);
\draw[flow] (tectonic) -- (pdf);

\node[steering, below=9mm of md] (template) {\texttt{templates/*.tex}};
\node[steering, below=9mm of pandoc] (filters) {\texttt{filters/*.lua}};
\draw[flow, dashed] (template) -- (pandoc);
\draw[flow, dashed] (filters) -- (pandoc);
\end{tikzpicture}
\caption{The two external programs, the files that steer pandoc, and the
intermediate \texttt{.tex} left on disk beside the source.\label{fig:pipeline}}
\end{figure}

## Goals

- Keep the source readable as plain text
- Let one metadata key change the entire look
- Fail with a message that names the offending line

# Failure Handling {#sec:failure}

Either external program can reject the document. The decision the builder makes
is shown in Figure \ref{fig:errors}.

\begin{figure}[H]
\centering
\begin{tikzpicture}[
  node distance = 8mm and 12mm,
  action/.style = {draw=accent, thick, rounded corners=2pt, fill=accent!5,
                 minimum width=24mm, minimum height=8mm, align=center,
                 font=\small, text=accent},
  choice/.style = {draw=accent, thick, diamond, aspect=2.2, fill=accent!5,
                   inner sep=1pt, align=center, font=\scriptsize, text=accent},
  done/.style = {draw=textgray!50, thick, rounded corners=2pt,
                 minimum width=20mm, minimum height=8mm, align=center,
                 font=\small, text=textgray},
  flow/.style = {-{Stealth[length=2mm]}, draw=accent, thick},
  edgetext/.style = {font=\scriptsize, text=textgray, inner sep=2pt}
]
\node[action] (run) {run the program};
\node[choice, below=of run] (ok) {exit 0?};
\node[done, below=12mm of ok] (next) {next stage};
\node[action, right=22mm of ok] (parse) {scan stderr};
\node[done, below=of parse] (report) {report the line};

\draw[flow] (run) -- (ok);
\draw[flow] (ok) -- node[edgetext, left] {yes} (next);
\draw[flow] (ok) -- node[edgetext, above] {no} (parse);
\draw[flow] (parse) -- (report);
\end{tikzpicture}
\caption{Error handling shared by both stages.\label{fig:errors}}
\end{figure}

A document that never reaches `tectonic` costs nothing to fail, which is why the
cheap check runs first.

# Measurements {#sec:measurements}

Build time splits unevenly between the two stages. Figure \ref{fig:timing} plots
a made up wall clock time for each stage, drawn entirely in TikZ without a
plotting package.

\begin{figure}[H]
\centering
\begin{tikzpicture}[y=0.10mm, x=1mm]
\draw[draw=textgray!30] (0,0) -- (78,0);
\foreach \y in {0, 200, 400, 600} {
  \draw[draw=textgray!30] (0,\y) -- (78,\y);
  \node[left, font=\scriptsize, text=textgray] at (0,\y) {\y};
}
\node[rotate=90, font=\scriptsize, text=textgray] at (-11,300) {milliseconds};

\foreach \i/\name/\pandoc/\tectonic in {
  0/minimal/40/210,
  1/tables/70/330,
  2/images/60/420,
  3/stress/130/615
} {
  \begin{scope}[xshift=\i*19mm]
    \fill[accent!35] (2,0) rectangle (8,\pandoc);
    \fill[accent] (9,0) rectangle (15,\tectonic);
    \node[below, font=\scriptsize, text=textgray] at (8.5,-2) {\name};
  \end{scope}
}

\fill[accent!35] (52,560) rectangle (56,590);
\node[right, font=\scriptsize, text=textgray] at (57,575) {pandoc};
\fill[accent] (52,500) rectangle (56,530);
\node[right, font=\scriptsize, text=textgray] at (57,515) {tectonic};
\end{tikzpicture}
\caption{Invented figures, shown to demonstrate a chart drawn in TikZ rather
than to report a measurement.\label{fig:timing}}
\end{figure}

The same invented numbers, for anyone who would rather read them:

| Document | pandoc (ms) | tectonic (ms) | Total (ms) |
| :------- | ----------: | ------------: | ---------: |
| minimal  |          40 |           210 |        250 |
| tables   |          70 |           330 |        400 |
| images   |          60 |           420 |        480 |
| stress   |         130 |           615 |        745 |

Raster images pass through the same figure machinery as the diagrams above, so
they can be referenced the same way. Figure \ref{fig:raster} is a PNG.

![A raster image, numbered alongside the drawn ones.\label{fig:raster}](image.png){width=40%}

# Cross References {#sec:crossrefs}

Both reference styles work in the same document.

Raw LaTeX, which numbers the target: Section \ref{sec:overview} introduces the
pipeline, Figure \ref{fig:errors} shows the error path, and Figure
\ref{fig:timing} appears on page \pageref{fig:timing}.

GFM anchors, which link by name: [the overview](#sec:overview),
[failure handling](#sec:failure), and [the goals](#goals).
