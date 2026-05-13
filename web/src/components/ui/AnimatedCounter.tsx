import { useEffect, useMemo, useRef, useState } from 'react';

import { clamp } from '@/lib/utils';

export function AnimatedCounter({ value, suffix = '' }: { value: number; suffix?: string }) {
  const [display, setDisplay] = useState(value);
  const previousValue = useRef(value);

  useEffect(() => {
    const start = performance.now();
    const initial = previousValue.current;
    const delta = value - initial;

    const frame = (time: number) => {
      const progress = clamp((time - start) / 500, 0, 1);
      const next = Math.round(initial + delta * progress);
      setDisplay(next);
      if (progress < 1) {
        requestAnimationFrame(frame);
      }
    };

    const id = requestAnimationFrame(frame);
    previousValue.current = value;
    return () => cancelAnimationFrame(id);
  }, [value]);

  const formatted = useMemo(() => new Intl.NumberFormat('ru-RU').format(display), [display]);

  return <span>{formatted}{suffix}</span>;
}
