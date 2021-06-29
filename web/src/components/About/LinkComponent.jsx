import { GitHub as GitHubIcon } from '@material-ui/icons'

import { LinkWrapper, LinkIcon } from './style'

export default function LinkComponent({ name, link }) {
  return (
    <LinkWrapper isLink={!!link} href={link} target='_blank' rel='noreferrer'>
      {link && (
        <LinkIcon>
          <GitHubIcon />
        </LinkIcon>
      )}

      <div>{name}</div>
    </LinkWrapper>
  )
}
