import { GitHub as GitHubIcon } from '@material-ui/icons'

import { NameWrapper, NameIcon } from './style'

export default function NameComponent({ name, link }) {
  return (
    <NameWrapper isLink={!!link} href={link} target='_blank' rel='noreferrer'>
      {link && (
        <NameIcon>
          <GitHubIcon />
        </NameIcon>
      )}

      {name}
    </NameWrapper>
  )
}
