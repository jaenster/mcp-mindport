'use client';

import ReactMarkdown from 'react-markdown';
import remarkGfm from 'remark-gfm';

interface MarkdownProps {
  content: string;
  className?: string;
}

export function Markdown({ content, className = '' }: MarkdownProps) {
  return (
    <div className={`prose prose-sm dark:prose-invert max-w-none ${className}`}>
      <ReactMarkdown
        remarkPlugins={[remarkGfm]}
        components={{
        // Customize code blocks
        code({ className, children, ...props }: any) {
          const match = /language-(\w+)/.exec(className || '');
          const isInline = !match;
          return isInline ? (
            <code
              className="bg-muted px-1.5 py-0.5 rounded text-sm font-mono"
              {...props}
            >
              {children}
            </code>
          ) : (
            <code
              className="block bg-muted p-4 rounded-md overflow-x-auto text-sm font-mono"
              {...props}
            >
              {children}
            </code>
          );
        },
        // Customize headings
        h1: ({ children }: any) => (
          <h1 className="text-2xl font-bold mb-4 border-b pb-2">{children}</h1>
        ),
        h2: ({ children }: any) => (
          <h2 className="text-xl font-semibold mb-3 mt-6">{children}</h2>
        ),
        h3: ({ children }: any) => (
          <h3 className="text-lg font-medium mb-2 mt-4">{children}</h3>
        ),
        // Customize links
        a: ({ href, children }: any) => (
          <a
            href={href}
            className="text-primary hover:underline"
            target="_blank"
            rel="noopener noreferrer"
          >
            {children}
          </a>
        ),
        // Customize tables
        table: ({ children }: any) => (
          <div className="overflow-x-auto">
            <table className="min-w-full border-collapse border border-border">
              {children}
            </table>
          </div>
        ),
        th: ({ children }: any) => (
          <th className="border border-border bg-muted px-4 py-2 text-left font-medium">
            {children}
          </th>
        ),
        td: ({ children }: any) => (
          <td className="border border-border px-4 py-2">{children}</td>
        ),
        // Customize blockquotes
        blockquote: ({ children }: any) => (
          <blockquote className="border-l-4 border-primary pl-4 italic text-muted-foreground">
            {children}
          </blockquote>
        ),
        // Customize lists
        ul: ({ children }: any) => (
          <ul className="list-disc list-inside space-y-1">{children}</ul>
        ),
        ol: ({ children }: any) => (
          <ol className="list-decimal list-inside space-y-1">{children}</ol>
        ),
        }}
      >
        {content}
      </ReactMarkdown>
    </div>
  );
}

export function detectContentType(content: string, mimeType?: string): 'markdown' | 'code' | 'text' {
  // Check mime type first
  if (mimeType) {
    if (mimeType.includes('markdown') || mimeType.includes('md')) {
      return 'markdown';
    }
    if (mimeType.startsWith('text/') && !mimeType.includes('plain')) {
      return 'code';
    }
  }

  // Check content patterns
  const markdownPatterns = [
    /^#{1,6}\s+.+$/m, // Headers
    /^\*\*.*\*\*$/m, // Bold
    /^-\s+.+$/m, // Lists
    /^\>\s+.+$/m, // Blockquotes
    /```[\s\S]*?```/m, // Code blocks
    /\[.*\]\(.*\)/m, // Links
  ];

  const codePatterns = [
    /^(import|export|function|class|const|let|var)\s+/m,
    /^\s*(def|class|if|for|while|return)\s+/m,
    /^<[^>]+>/m, // HTML tags
    /^\s*\{[\s\S]*\}$/m, // JSON
  ];

  const markdownScore = markdownPatterns.reduce(
    (score, pattern) => score + (pattern.test(content) ? 1 : 0),
    0
  );

  const codeScore = codePatterns.reduce(
    (score, pattern) => score + (pattern.test(content) ? 1 : 0),
    0
  );

  if (markdownScore > codeScore && markdownScore > 0) {
    return 'markdown';
  }
  if (codeScore > 0) {
    return 'code';
  }
  return 'text';
}