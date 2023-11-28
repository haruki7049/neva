// Code generated by tygo. DO NOT EDIT.
import * as src from "./sourcecode";
//////////
// source: resolve_file.go

export interface ResolveFileRequest {
  workspaceUri: URI;
  document: {
    uri: URI;
    fileName: string;
  };
}
export interface URI {
  path: string;
  fsPath: string;
}
export interface ResolveFileResponce {
  file: src.File;
  extra: Extra; // info that is not presented in the file but needed for rendering
}
export interface Extra {
  nodesPorts: {
    [key: string]: { [key: string]: src.Interface };
  }; // components -> nodes -> interface
}
