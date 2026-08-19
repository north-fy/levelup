import { useTheme } from '../../theme'
import Switch from '../ui/Switch'

export default function ThemeToggle() {
  const { theme, toggle } = useTheme()
  const dark = theme === 'dark'
  return (
    <span className="inline-flex items-center gap-2">
      <span aria-hidden="true" className="text-sm">
        {dark ? '🌙' : '☀️'}
      </span>
      <Switch checked={dark} onChange={toggle} label={dark ? 'Выключить тёмную тему' : 'Включить тёмную тему'} />
    </span>
  )
}