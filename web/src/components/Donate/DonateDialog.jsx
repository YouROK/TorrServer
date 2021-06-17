import ListItem from '@material-ui/core/ListItem'
import Dialog from '@material-ui/core/Dialog'
import DialogTitle from '@material-ui/core/DialogTitle'
import DialogContent from '@material-ui/core/DialogContent'
import DialogActions from '@material-ui/core/DialogActions'
import List from '@material-ui/core/List'
import ButtonGroup from '@material-ui/core/ButtonGroup'
import Button from '@material-ui/core/Button'
import { useTranslation } from 'react-i18next'

const donateFrame =
  '<iframe src="https://yoomoney.ru/quickpay/shop-widget?writer=seller&targets=TorrServer Donate&targets-hint=&default-sum=200&button-text=14&payment-type-choice=on&mobile-payment-type-choice=on&comment=on&hint=&successURL=&quickpay=shop&account=410013733697114" width="320" height="320" frameborder="0" allowtransparency="true" scrolling="no"></iframe>'

export default function DonateDialog({ onClose }) {
  const { t } = useTranslation()

  return (
    <Dialog open onClose={onClose} aria-labelledby='form-dialog-title' fullWidth maxWidth='xs'>
      <DialogTitle id='form-dialog-title'>{t('Donate')}</DialogTitle>
      <DialogContent>
        <List>
          <ListItem key='DonateLinks'>
            <ButtonGroup variant='outlined' color='primary' aria-label='contained primary button group'>
              <Button onClick={() => window.open('https://www.paypal.com/paypalme/yourok', '_blank')}>PayPal</Button>
              <Button onClick={() => window.open('https://yoomoney.ru/to/410013733697114', '_blank')}>IO.Money</Button>
            </ButtonGroup>
          </ListItem>
          <ListItem key='DonateForm'>
            {/* eslint-disable-next-line react/no-danger */}
            <div dangerouslySetInnerHTML={{ __html: donateFrame }} />
          </ListItem>
        </List>
      </DialogContent>

      <DialogActions>
        <Button onClick={onClose} color='primary' variant='outlined'>
          Ok
        </Button>
      </DialogActions>
    </Dialog>
  )
}
