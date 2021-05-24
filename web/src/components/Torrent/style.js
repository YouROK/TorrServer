import styled, { css } from 'styled-components';

export const TorrentCard = styled.div`
    border: 1px solid;
    border-radius: 5px;
    display: grid;
    grid-template-columns: repeat(2, 1fr);
    grid-template-rows: 175px minmax(min-content, 1fr);
    grid-template-areas:
        "poster buttons"
        "description description";
    gap: 10px;
    padding: 10px;
    background: #3fb57a;
    box-shadow:
        0px 2px 4px -1px rgb(0 0 0 / 20%),
        0px 4px 5px 0px rgb(0 0 0 / 14%),
        0px 1px 10px 0px rgb(0 0 0 / 12%);
`

export const TorrentCardPoster = styled.div`
    grid-area: poster;
    border-radius: 5px;
    overflow: hidden;
    text-align: center;

    ${({ isPoster }) => isPoster ? css`
        img {
            height: 100%;
            border-radius: 5px;
        }
    `: css`
        display: grid;
        place-items: center;
        background: #74c39c;
        border: 1px solid; 

        svg {
            transform: translateY(-3px);
        }
    `};
`
export const TorrentCardButtons = styled.div`
    grid-area: buttons;
    display: grid;
    gap: 5px;
`
export const TorrentCardDescription = styled.div`
    grid-area: description;
    background: #74c39c;
    border-radius: 5px;
    padding: 5px;
    word-break: break-word;
`

export const TorrentCardDescriptionLabel = styled.div`
    text-transform: uppercase;
    font-size: 10px;
    font-weight: 500;
    letter-spacing: 0.4px;
    color: #216e47;
`

export const TorrentCardDescriptionContent = styled.div`
    margin-left: 5px;
    margin-bottom: 10px;
`

export const StyledButton = styled.button`
    border-radius: 5px;
    border: none;
    cursor: pointer;
    transition: 0.2s;
    display: flex;
    align-items: center;
    text-transform: uppercase;
    background: #216e47;
    color: #fff;
    font-size: 1rem;
    font-family: "Roboto", "Helvetica", "Arial", sans-serif;
    letter-spacing: 0.009em;

    > :first-child {
        margin-right: 10px;
    }

    @media (max-width: 600px) {
        font-size: 0.7rem;

        > :first-child {
            margin-right: 15px;
        }
    }


    :hover {
        background: #2a7e54;
    }
`