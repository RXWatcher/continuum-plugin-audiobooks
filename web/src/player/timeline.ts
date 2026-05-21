import type { AudiobookChapter, AudiobookFile } from '@/api/types';

export type FileTimelineEntry = {
  fileIndex: number;
  fileOrdinal: number;
  start: number;
  end: number;
  duration: number;
};

export type FileTimePosition = {
  fileIndex: number;
  fileOrdinal: number;
  fileTime: number;
  fileStart: number;
};

export function buildFileTimeline(files: AudiobookFile[]): FileTimelineEntry[] {
  let start = 0;
  return files.map((file, fileOrdinal) => {
    const duration = Math.max(0, file.duration_seconds || 0);
    const entry = {
      fileIndex: file.index,
      fileOrdinal,
      start,
      end: start + duration,
      duration,
    };
    start += duration;
    return entry;
  });
}

export function timelineDuration(timeline: FileTimelineEntry[]): number {
  return timeline.at(-1)?.end ?? 0;
}

export function positionToFileTime(
  timeline: FileTimelineEntry[],
  bookTime: number,
): FileTimePosition {
  if (timeline.length === 0) {
    return { fileIndex: 0, fileOrdinal: 0, fileTime: 0, fileStart: 0 };
  }
  const duration = timelineDuration(timeline);
  const clamped = Math.max(0, Math.min(bookTime, duration));
  const entry =
    timeline.find((item) => clamped >= item.start && clamped < item.end) ??
    timeline[timeline.length - 1];
  return {
    fileIndex: entry.fileIndex,
    fileOrdinal: entry.fileOrdinal,
    fileTime: Math.max(0, Math.min(clamped - entry.start, entry.duration)),
    fileStart: entry.start,
  };
}

export function fileTimeToBookTime(
  timeline: FileTimelineEntry[],
  fileOrdinal: number,
  fileTime: number,
): number {
  const entry = timeline[fileOrdinal] ?? timeline[0];
  if (!entry) return 0;
  return entry.start + Math.max(0, Math.min(fileTime, entry.duration));
}

// chapterAt picks the chapter containing bookTime. Three buckets, in
// priority order:
//   1. Exact match — bookTime falls inside [start, end).
//   2. Pre-first — bookTime sits before chapters[0].start (intro music,
//      or chapters that don't start at zero). Return chapters[0], not
//      chapters[last] — the previous fallback was a footgun: returning the
//      last chapter as "current" made the next-chapter button silently
//      disable itself at the start of every book whose chapters[0] doesn't
//      start at second zero.
//   3. Gap or trailing — bookTime sits between two chapters (gap in the
//      timeline) or after the last chapter's end. Return the most-recent
//      chapter whose start_seconds <= bookTime, falling back to the last
//      chapter when bookTime is past everything.
export function chapterAt(
  chapters: AudiobookChapter[] | undefined,
  bookTime: number,
): AudiobookChapter | undefined {
  if (!chapters?.length) return undefined;
  // Exact match.
  for (const c of chapters) {
    if (bookTime >= c.start_seconds && bookTime < c.end_seconds) return c;
  }
  // Pre-first.
  if (bookTime < chapters[0].start_seconds) return chapters[0];
  // Gap or trailing — walk forward until we pass bookTime, then take the
  // previous chapter as the "current". Assumes chapters are in
  // chronological order, which matches ABS contract.
  let last = chapters[0];
  for (const c of chapters) {
    if (c.start_seconds > bookTime) return last;
    last = c;
  }
  return last;
}
