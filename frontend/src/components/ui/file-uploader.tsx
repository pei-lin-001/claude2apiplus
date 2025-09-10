import React, { useState, useCallback } from 'react';
import { useDropzone } from 'react-dropzone';
import { Upload, X, FileImage } from 'lucide-react';
import { Button } from './button';
import { formatFileSize } from '../../utils';
import { FileUpload } from '../../types';

interface FileUploaderProps {
  onFilesChange: (files: FileUpload[]) => void;
  maxFiles?: number;
  maxSize?: number; // in bytes
  acceptedTypes?: string[];
  className?: string;
}

const FileUploader: React.FC<FileUploaderProps> = ({
  onFilesChange,
  maxFiles = 5,
  maxSize = 10 * 1024 * 1024, // 10MB
  acceptedTypes = ['image/*'],
  className = '',
}) => {
  const [files, setFiles] = useState<FileUpload[]>([]);

  const onDrop = useCallback((acceptedFiles: File[]) => {
    const newFiles: FileUpload[] = [];
    
    acceptedFiles.forEach((file) => {
      const reader = new FileReader();
      reader.onload = (e) => {
        const result = e.target?.result as string;
        const fileUpload: FileUpload = {
          id: crypto.randomUUID(),
          name: file.name,
          type: file.type,
          size: file.size,
          url: result,
        };
        newFiles.push(fileUpload);
        
        if (newFiles.length === acceptedFiles.length) {
          const updatedFiles = [...files, ...newFiles].slice(0, maxFiles);
          setFiles(updatedFiles);
          onFilesChange(updatedFiles);
        }
      };
      reader.readAsDataURL(file);
    });
  }, [files, maxFiles, onFilesChange]);

  const { getRootProps, getInputProps, isDragActive } = useDropzone({
    onDrop,
    accept: acceptedTypes.reduce((acc, type) => ({ ...acc, [type]: [] }), {}),
    maxSize,
    multiple: maxFiles > 1,
    disabled: files.length >= maxFiles,
  });

  const removeFile = (fileId: string) => {
    const updatedFiles = files.filter(file => file.id !== fileId);
    setFiles(updatedFiles);
    onFilesChange(updatedFiles);
  };

  return (
    <div className={`space-y-4 ${className}`}>
      {/* 上传区域 */}
      {files.length < maxFiles && (
        <div
          {...getRootProps()}
          className={`border-2 border-dashed rounded-lg p-6 text-center cursor-pointer transition-colors ${
            isDragActive
              ? 'border-primary bg-primary/5'
              : 'border-border hover:border-primary/50 hover:bg-accent/5'
          }`}
        >
          <input {...getInputProps()} />
          <Upload className="w-12 h-12 mx-auto mb-4 text-muted-foreground" />
          <div className="space-y-2">
            <p className="text-lg font-medium">
              {isDragActive ? '松开以上传文件' : '拖拽文件到这里'}
            </p>
            <p className="text-sm text-muted-foreground">
              或点击选择文件
            </p>
            <p className="text-xs text-muted-foreground">
              支持 {acceptedTypes.join(', ')} 格式，最大 {formatFileSize(maxSize)}
            </p>
          </div>
        </div>
      )}

      {/* 文件列表 */}
      {files.length > 0 && (
        <div className="space-y-2">
          <div className="flex items-center justify-between text-sm font-medium">
            <span>已选择文件 ({files.length}/{maxFiles})</span>
            <Button
              variant="ghost"
              size="sm"
              onClick={() => {
                setFiles([]);
                onFilesChange([]);
              }}
            >
              清空全部
            </Button>
          </div>
          
          <div className="grid grid-cols-1 gap-2">
            {files.map((file) => (
              <div
                key={file.id}
                className="flex items-center justify-between p-3 bg-muted rounded-lg group"
              >
                <div className="flex items-center space-x-3 flex-1 min-w-0">
                  <FileImage className="w-5 h-5 text-muted-foreground flex-shrink-0" />
                  <div className="flex-1 min-w-0">
                    <p className="text-sm font-medium truncate">{file.name}</p>
                    <p className="text-xs text-muted-foreground">
                      {formatFileSize(file.size)}
                    </p>
                  </div>
                </div>
                
                <Button
                  variant="ghost"
                  size="icon"
                  className="opacity-0 group-hover:opacity-100 transition-opacity flex-shrink-0"
                  onClick={() => removeFile(file.id)}
                >
                  <X className="w-4 h-4" />
                </Button>
              </div>
            ))}
          </div>
        </div>
      )}
    </div>
  );
};

export default FileUploader;