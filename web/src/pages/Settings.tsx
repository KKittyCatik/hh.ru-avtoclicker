import * as Select from '@radix-ui/react-select';
import * as Slider from '@radix-ui/react-slider';
import * as Switch from '@radix-ui/react-switch';
import * as Tabs from '@radix-ui/react-tabs';
import { Check, ChevronDown } from 'lucide-react';

import { GlowCard } from '@/components/ui/GlowCard';
import { useAuthStore } from '@/store/auth';

export default function SettingsPage() {
  const settings = useAuthStore((state) => state.settings);
  const updateSettings = useAuthStore((state) => state.updateSettings);

  return (
    <Tabs.Root defaultValue="automation" className="space-y-4">
      <Tabs.List className="glass-panel inline-flex rounded-2xl p-1">
        {['automation', 'filters'].map((value) => (
          <Tabs.Trigger key={value} className="rounded-2xl px-4 py-2 text-sm text-muted data-[state=active]:bg-white/8 data-[state=active]:text-foreground" value={value}>
            {value === 'automation' ? 'Automation' : 'Filters'}
          </Tabs.Trigger>
        ))}
      </Tabs.List>

      <Tabs.Content value="automation">
        <GlowCard className="space-y-6">
          <div>
            <p className="mb-2 text-sm text-muted">LLM provider</p>
            <Select.Root value={settings.llmProvider} onValueChange={(value: 'openai' | 'deepseek') => updateSettings({ llmProvider: value })}>
              <Select.Trigger className="flex w-full items-center justify-between rounded-2xl border border-white/10 bg-white/[0.03] px-4 py-3 text-sm text-foreground">
                <Select.Value />
                <ChevronDown className="h-4 w-4 text-muted" />
              </Select.Trigger>
              <Select.Portal>
                <Select.Content className="z-50 overflow-hidden rounded-2xl border border-white/10 bg-[#0d0d11] p-1 shadow-soft">
                  <Select.Viewport>
                    {['openai', 'deepseek'].map((value) => (
                      <Select.Item key={value} className="flex cursor-pointer items-center justify-between rounded-xl px-3 py-2 text-sm text-foreground outline-none hover:bg-white/5" value={value}>
                        <Select.ItemText>{value}</Select.ItemText>
                        <Select.ItemIndicator><Check className="h-4 w-4" /></Select.ItemIndicator>
                      </Select.Item>
                    ))}
                  </Select.Viewport>
                </Select.Content>
              </Select.Portal>
            </Select.Root>
          </div>

          <div>
            <div className="mb-2 flex items-center justify-between text-sm text-muted">
              <span>Daily limit</span>
              <span className="text-foreground">{settings.dailyLimit}</span>
            </div>
            <Slider.Root className="relative flex h-6 w-full touch-none items-center" max={150} min={10} step={5} value={[settings.dailyLimit]} onValueChange={([value]) => updateSettings({ dailyLimit: value })}>
              <Slider.Track className="relative h-2 grow rounded-full bg-white/10">
                <Slider.Range className="absolute h-full rounded-full bg-gradient-to-r from-violet-600 to-cyan-500" />
              </Slider.Track>
              <Slider.Thumb className="block h-5 w-5 rounded-full border border-white/20 bg-white shadow" />
            </Slider.Root>
          </div>

          <div className="flex items-center justify-between rounded-2xl border border-white/10 bg-white/[0.03] p-4">
            <div>
              <p className="text-sm font-medium text-foreground">Exclude agencies</p>
              <p className="text-sm text-muted">Скрывать кадровые агентства и массовый рекрутинг.</p>
            </div>
            <Switch.Root checked={settings.excludeAgencies} className="relative h-7 w-12 rounded-full bg-white/10 data-[state=checked]:bg-violet-600" onCheckedChange={(value) => updateSettings({ excludeAgencies: value })}>
              <Switch.Thumb className="block h-5 w-5 translate-x-1 rounded-full bg-white transition data-[state=checked]:translate-x-6" />
            </Switch.Root>
          </div>
        </GlowCard>
      </Tabs.Content>

      <Tabs.Content value="filters">
        <GlowCard className="space-y-6">
          <div>
            <p className="mb-2 text-sm text-muted">Schedule filter</p>
            <Select.Root value={settings.scheduleFilter} onValueChange={(value: 'remote' | 'hybrid' | 'office' | 'any') => updateSettings({ scheduleFilter: value })}>
              <Select.Trigger className="flex w-full items-center justify-between rounded-2xl border border-white/10 bg-white/[0.03] px-4 py-3 text-sm text-foreground">
                <Select.Value />
                <ChevronDown className="h-4 w-4 text-muted" />
              </Select.Trigger>
              <Select.Portal>
                <Select.Content className="z-50 overflow-hidden rounded-2xl border border-white/10 bg-[#0d0d11] p-1 shadow-soft">
                  <Select.Viewport>
                    {['remote', 'hybrid', 'office', 'any'].map((value) => (
                      <Select.Item key={value} className="flex cursor-pointer items-center justify-between rounded-xl px-3 py-2 text-sm text-foreground outline-none hover:bg-white/5" value={value}>
                        <Select.ItemText>{value}</Select.ItemText>
                        <Select.ItemIndicator><Check className="h-4 w-4" /></Select.ItemIndicator>
                      </Select.Item>
                    ))}
                  </Select.Viewport>
                </Select.Content>
              </Select.Portal>
            </Select.Root>
          </div>
          <div>
            <div className="mb-2 flex items-center justify-between text-sm text-muted">
              <span>Minimum salary</span>
              <span className="text-foreground">{settings.minimumSalary.toLocaleString('ru-RU')} ₽</span>
            </div>
            <Slider.Root className="relative flex h-6 w-full touch-none items-center" max={500000} min={50000} step={10000} value={[settings.minimumSalary]} onValueChange={([value]) => updateSettings({ minimumSalary: value })}>
              <Slider.Track className="relative h-2 grow rounded-full bg-white/10">
                <Slider.Range className="absolute h-full rounded-full bg-gradient-to-r from-violet-600 to-cyan-500" />
              </Slider.Track>
              <Slider.Thumb className="block h-5 w-5 rounded-full border border-white/20 bg-white shadow" />
            </Slider.Root>
          </div>
        </GlowCard>
      </Tabs.Content>
    </Tabs.Root>
  );
}
