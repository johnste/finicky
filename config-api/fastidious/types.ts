export interface ISchema {
  [index: string]: IValidator | number | string;
}

export interface IValidator {
  (value: any, key: string): string | undefined;
  typeName: NameType;
}

export type ICheckType = IValidator & {
  isRequired: IValidator;
};

export type TypeCallback = (value: any, key: string) => string[] | boolean;

interface IStructuredName {
  [index: string]: string | IStructuredName;
}

export type NameType = string | IStructuredName;
