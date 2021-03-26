import React from 'react';
import PropTypes from 'prop-types';

import "./style.scss"

const { formatText, messageHtmlToComponent } = window.PostUtils;


export default class SpoilerPost extends React.PureComponent {
    static propTypes = {
        post: PropTypes.object.isRequired,
        theme: PropTypes.object.isRequired,
    }

    render() {
        const post = this.props.post;
        if (post == null) return null;
        const props = { ...(post.props || {}) };
        const spoiler_text = messageHtmlToComponent(formatText(props.spoiler_text || ''));
        return <div className="spoiler--wrapper">{spoiler_text}</div>;
    }
}

