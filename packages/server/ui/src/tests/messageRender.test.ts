import {describe, expect, it} from 'vitest';
import {contentType, RenderMode} from '../message/extras';

describe('message contentType', () => {
    it('renders Markdown when no extra is set', () =>
        expect(contentType(undefined)).toBe(RenderMode.Markdown));
    it('honours an explicit markdown extra', () =>
        expect(contentType({'client::display': {contentType: 'text/markdown'}})).toBe(
            RenderMode.Markdown
        ));
    it('honours an explicit plain-text extra', () =>
        expect(contentType({'client::display': {contentType: 'text/plain'}})).toBe(
            RenderMode.Plain
        ));
});
