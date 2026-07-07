import { Card } from '@/components/ui/card';
import { cn } from '@/lib/utils';

export type Column<T> = {
  key: string;
  header: string;
  cell: (row: T) => React.ReactNode;
  className?: string;
};

export function DataTable<T>({ columns, data }: { columns: Column<T>[]; data: T[] }) {
  return (
    <Card className="overflow-hidden">
      <div className="overflow-x-auto">
        <table className="w-full min-w-[760px] text-left text-sm">
          <thead className="bg-slate-900/80 text-xs uppercase tracking-widest text-slate-500">
            <tr>
              {columns.map((column) => (
                <th key={column.key} className={cn('px-4 py-3 font-medium', column.className)}>{column.header}</th>
              ))}
            </tr>
          </thead>
          <tbody className="divide-y divide-slate-800">
            {data.map((row, index) => (
              <tr key={index} className="bg-slate-950/40 text-slate-300 transition hover:bg-slate-900/70">
                {columns.map((column) => (
                  <td key={column.key} className={cn('px-4 py-4', column.className)}>{column.cell(row)}</td>
                ))}
              </tr>
            ))}
          </tbody>
        </table>
      </div>
    </Card>
  );
}
