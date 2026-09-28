/**
 * @jest-environment jsdom
 */

import '@testing-library/jest-dom';
import {fireEvent, render, screen} from '@testing-library/react';
import React from 'react';

import type {Post} from '@mattermost/types/posts';

import SpoilerPost, {SPOILER_POST_TYPE} from './spoiler_post';

function makePost(spoilerText?: string): Post {
    return {
        id: 'post1',
        type: SPOILER_POST_TYPE,
        message: 'Spoiler — hidden until revealed',
        props: spoilerText === undefined ? {} : {spoiler_text: spoilerText},
    } as unknown as Post;
}

beforeEach(() => {
    window.PostUtils = {
        formatText: (text: string) => text,
        messageHtmlToComponent: (html: string) => <span>{html}</span>,
    };
});

test('hides the spoiler until it is revealed', () => {
    render(<SpoilerPost post={makePost('The butler did it')}/>);

    const content = screen.getByText('The butler did it').parentElement!;
    expect(content).toHaveAttribute('aria-hidden', 'true');
    expect(content).toHaveAttribute('inert');

    fireEvent.click(screen.getByRole('button', {name: 'Reveal spoiler'}));

    expect(content).toHaveAttribute('aria-hidden', 'false');
    expect(content).not.toHaveAttribute('inert');
    expect(screen.getByTestId('spoilerPost')).toHaveClass('spoiler-post--revealed');
});

test('can hide the spoiler again', () => {
    render(<SpoilerPost post={makePost('The butler did it')}/>);

    fireEvent.click(screen.getByRole('button', {name: 'Reveal spoiler'}));
    fireEvent.click(screen.getByRole('button', {name: 'Hide spoiler'}));

    expect(screen.getByRole('button', {name: 'Reveal spoiler'})).toBeInTheDocument();
    expect(screen.getByTestId('spoilerPost')).not.toHaveClass('spoiler-post--revealed');
});

test('falls back to the post message when there is no spoiler text', () => {
    render(<SpoilerPost post={makePost()}/>);

    expect(screen.getByText('Spoiler — hidden until revealed')).toBeInTheDocument();
    expect(screen.queryByRole('button')).not.toBeInTheDocument();
});
