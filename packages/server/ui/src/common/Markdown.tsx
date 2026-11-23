import React from 'react';
import ReactMarkdown, {defaultUrlTransform} from 'react-markdown';
import type {UrlTransform} from 'react-markdown';
import gfm from 'remark-gfm';
import breaks from 'remark-breaks';
import Button from '@mui/material/Button';
import OpenInNew from '@mui/icons-material/OpenInNew';

// Copy from mlflow/server/js/src/shared/web-shared/genai-markdown-renderer/GenAIMarkdownRenderer.tsx
// Related PR: https://github.com/mlflow/mlflow/pull/16761
const urlTransform: UrlTransform = (value) => {
    if (
        value.startsWith('data:image/png;') ||
        value.startsWith('data:image/jpeg;') ||
        value.startsWith('data:image/gif;')
    ) {
        return value;
    }
    return defaultUrlTransform(value);
};

// Any Markdown link becomes a button. Messages that only contain text have no
// button; this is driven by the content, not by a specific app or URL.
const LinkButton = ({href, children}: {href?: string; children?: React.ReactNode}) => {
    if (!href) {
        return <span>{children}</span>;
    }
    return (
        <Button
            component="a"
            href={href}
            target="_blank"
            rel="noopener noreferrer"
            variant="outlined"
            size="small"
            endIcon={<OpenInNew />}
            style={{margin: '4px 0'}}>
            {children}
        </Button>
    );
};

export const Markdown = ({
    children,
    onImageLoaded = () => {},
}: {
    children: string;
    onImageLoaded?: () => void;
}) => (
    <ReactMarkdown
        components={{
            img: ({...props}) => <img onLoad={onImageLoaded} {...props} />,
            a: ({href, children}) => <LinkButton href={href}>{children}</LinkButton>,
        }}
        remarkPlugins={[gfm, breaks]}
        urlTransform={urlTransform}>
        {children}
    </ReactMarkdown>
);
