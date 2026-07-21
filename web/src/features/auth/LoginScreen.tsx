import { useState } from 'react'
import axios from 'axios'
import { Button, Input, Label, Spinner, TextField } from '@heroui/react'
import { LogIn } from 'lucide-react'
import { useTranslation } from 'react-i18next'

import { encodeBasicAuthorization, setCredentials } from 'shared/api/authCredentials'
import { torrentsHost } from 'shared/api/hosts'
import { publicUrl } from 'shared/lib/publicUrl'
import { iconMenu } from 'shared/ui/iconProps'

export interface LoginScreenProps {
  onSuccess: () => void
}

/** Branded Basic Auth form for the web UI only (accs.db + --httpauth). */
export default function LoginScreen({ onSuccess }: LoginScreenProps) {
  const { t } = useTranslation()
  const [username, setUsername] = useState('')
  const [password, setPassword] = useState('')
  const [error, setError] = useState<string | null>(null)
  const [pending, setPending] = useState(false)

  const submit = async () => {
    const user = username.trim()
    if (!user) {
      setError(t('LoginFailed'))
      return
    }
    setPending(true)
    setError(null)
    const authorization = encodeBasicAuthorization(user, password)
    try {
      const response = await axios.post(
        torrentsHost(),
        { action: 'list' },
        {
          headers: {
            Accept: 'application/json',
            'X-Requested-With': 'XMLHttpRequest',
            Authorization: authorization,
          },
          validateStatus: status => status === 200 || status === 401 || status === 403,
        },
      )
      if (response.status !== 200) {
        setError(t('LoginFailed'))
        return
      }
      setCredentials(user, password)
      onSuccess()
    } catch {
      setError(t('LoginFailed'))
    } finally {
      setPending(false)
    }
  }

  return (
    <div className='relative flex min-h-0 flex-1 flex-col items-center justify-center overflow-y-auto bg-background px-4 py-[max(2.5rem,env(safe-area-inset-top,0px))] pb-[max(2.5rem,env(safe-area-inset-bottom,0px))]'>
      <div
        className='pointer-events-none absolute inset-0 opacity-80'
        style={{
          background:
            'radial-gradient(ellipse 80% 50% at 50% -20%, color-mix(in oklab, var(--accent) 28%, transparent), transparent), radial-gradient(ellipse 60% 40% at 100% 100%, color-mix(in oklab, var(--accent) 12%, transparent), transparent)',
        }}
        aria-hidden
      />
      <form
        className='relative w-full max-w-sm rounded-2xl border border-border bg-surface/95 p-6 shadow-xl backdrop-blur'
        onSubmit={event => {
          event.preventDefault()
          if (!pending) void submit()
        }}
      >
        <div className='mb-6 flex flex-col items-center gap-3 text-center'>
          <img src={publicUrl('icon.png')} alt='' className='size-16 rounded-2xl shadow-lg' />
          <div>
            <h1 className='text-xl font-semibold text-foreground'>TorrServer</h1>
            <p className='mt-1 text-sm text-muted'>{t('LoginSubtitle')}</p>
          </div>
        </div>

        <div className='flex flex-col gap-4'>
          <TextField value={username} onChange={setUsername} isRequired isDisabled={pending} autoComplete='username'>
            <Label>{t('Username')}</Label>
            <Input autoFocus className='min-h-11' name='username' />
          </TextField>
          <TextField value={password} onChange={setPassword} isDisabled={pending} autoComplete='current-password'>
            <Label>{t('Password')}</Label>
            <Input type='password' className='min-h-11' name='password' />
          </TextField>

          {error ? (
            <p className='text-sm text-danger' role='alert'>
              {error}
            </p>
          ) : null}

          <Button type='submit' variant='primary' className='min-h-11 w-full gap-2' isPending={pending}>
            {({ isPending }) => (
              <>
                {isPending ? <Spinner size='sm' color='current' /> : <LogIn {...iconMenu} aria-hidden />}
                {t('SignIn')}
              </>
            )}
          </Button>
        </div>
      </form>
    </div>
  )
}
