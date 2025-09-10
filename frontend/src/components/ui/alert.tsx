import React from 'react';
import { AlertCircle, CheckCircle, XCircle, Info } from 'lucide-react';
import { cn } from '../../utils';

export interface AlertProps {
  variant?: 'default' | 'destructive' | 'success' | 'warning';
  className?: string;
  children: React.ReactNode;
}

const Alert: React.FC<AlertProps> = ({ 
  variant = 'default', 
  className, 
  children 
}) => {
  const variantClasses = {
    default: 'bg-background text-foreground border',
    destructive: 'bg-destructive text-destructive-foreground border-destructive',
    success: 'bg-green-50 text-green-700 dark:bg-green-950 dark:text-green-300 border-green-200 dark:border-green-800',
    warning: 'bg-yellow-50 text-yellow-700 dark:bg-yellow-950 dark:text-yellow-300 border-yellow-200 dark:border-yellow-800',
  };

  const icons = {
    default: Info,
    destructive: XCircle,
    success: CheckCircle,
    warning: AlertCircle,
  };

  const Icon = icons[variant];

  return (
    <div
      className={cn(
        'relative w-full rounded-lg border p-4',
        variantClasses[variant],
        className
      )}
    >
      <div className="flex items-start space-x-3">
        <Icon className="h-5 w-5 mt-0.5 flex-shrink-0" />
        <div className="flex-1">
          {children}
        </div>
      </div>
    </div>
  );
};

export default Alert;