export default function Spinner({ size = 'md', className = '' }) {
  const sz = { sm: 'w-4 h-4', md: 'w-8 h-8', lg: 'w-12 h-12' }[size];
  return (
    <div className={`${sz} ${className} animate-spin rounded-full border-2 border-gray-200 border-t-indigo-600`} />
  );
}
