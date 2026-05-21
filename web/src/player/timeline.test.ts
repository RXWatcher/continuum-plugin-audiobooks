import { describe, expect, test } from 'vitest';
import {
  buildFileTimeline,
  chapterAt,
  fileTimeToBookTime,
  positionToFileTime,
} from './timeline';

const files = [
  { index: 7, mime_type: 'audio/mpeg', format: 'mp3', duration_seconds: 100 },
  { index: 9, mime_type: 'audio/mpeg', format: 'mp3', duration_seconds: 50 },
  { index: 10, mime_type: 'audio/mpeg', format: 'mp3', duration_seconds: 75 },
];

describe('audiobook whole-book timeline', () => {
  test('maps whole-book seconds to the active file and local file time', () => {
    const timeline = buildFileTimeline(files);

    expect(positionToFileTime(timeline, 0)).toEqual({
      fileIndex: 7,
      fileOrdinal: 0,
      fileTime: 0,
      fileStart: 0,
    });
    expect(positionToFileTime(timeline, 125)).toEqual({
      fileIndex: 9,
      fileOrdinal: 1,
      fileTime: 25,
      fileStart: 100,
    });
    expect(positionToFileTime(timeline, 225)).toEqual({
      fileIndex: 10,
      fileOrdinal: 2,
      fileTime: 75,
      fileStart: 150,
    });
  });

  test('maps file-local time back to whole-book seconds', () => {
    const timeline = buildFileTimeline(files);

    expect(fileTimeToBookTime(timeline, 0, 12)).toBe(12);
    expect(fileTimeToBookTime(timeline, 1, 12)).toBe(112);
    expect(fileTimeToBookTime(timeline, 2, 100)).toBe(225);
  });

  test('derives the active chapter from whole-book seconds', () => {
    const chapters = [
      { title: 'Opening', start_seconds: 0, end_seconds: 45 },
      { title: 'Middle', start_seconds: 45, end_seconds: 120 },
      { title: 'Ending', start_seconds: 120, end_seconds: 225 },
    ];

    expect(chapterAt(chapters, 44)?.title).toBe('Opening');
    expect(chapterAt(chapters, 45)?.title).toBe('Middle');
    expect(chapterAt(chapters, 224)?.title).toBe('Ending');
    expect(chapterAt(chapters, 225)?.title).toBe('Ending');
  });

  test('returns the first chapter when bookTime precedes chapters[0]', () => {
    // Some books open with intro music or other content that sits before
    // the first chapter marker. The previous fallback returned
    // chapters[chapters.length-1], which made the next-chapter button
    // disable itself at the start of every such book.
    const chapters = [
      { title: 'Opening', start_seconds: 30, end_seconds: 90 },
      { title: 'Middle', start_seconds: 90, end_seconds: 180 },
    ];
    expect(chapterAt(chapters, 0)?.title).toBe('Opening');
    expect(chapterAt(chapters, 29.9)?.title).toBe('Opening');
  });

  test('returns the most-recent chapter when bookTime sits in a gap', () => {
    // Two chapters with a gap between them — a timeline glitch but real
    // ABS feeds occasionally have this. chapterAt should report the most
    // recently entered chapter rather than the last one in the array.
    const chapters = [
      { title: 'One', start_seconds: 0, end_seconds: 30 },
      { title: 'Two', start_seconds: 60, end_seconds: 120 },
    ];
    expect(chapterAt(chapters, 45)?.title).toBe('One');
    expect(chapterAt(chapters, 200)?.title).toBe('Two');
  });
});
