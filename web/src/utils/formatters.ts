export type Numberish = number | string;

export const toNumber = (value?: Numberish | null): number | undefined => {
  if (value === undefined || value === null || value === '') {
    return undefined;
  }

  const numeric = typeof value === 'number' ? value : Number(value);
  return Number.isFinite(numeric) ? numeric : undefined;
};

export const sumNumberish = (...values: Array<Numberish | null | undefined>): number =>
  values.reduce<number>((total, value) => total + (toNumber(value) ?? 0), 0);

export const formatBytes = (bytes?: Numberish | null): string => {
  const numeric = toNumber(bytes);
  if (numeric === undefined) {
    return '-';
  }
  if (numeric === 0) {
    return '0 B';
  }

  const units = ['B', 'KB', 'MB', 'GB', 'TB'];
  let value = numeric;
  let index = 0;
  while (Math.abs(value) >= 1024 && index < units.length - 1) {
    value /= 1024;
    index += 1;
  }

  return `${parseFloat(value.toFixed(2))} ${units[index]}`;
};
