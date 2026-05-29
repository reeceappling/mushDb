import {Note} from "@/app/components/formSubcomponents/notes";
import {EntryPerms, IsValidEntryPermSubset, IsValidUserIdPair} from "@/app/components/perms";
import {IsBool, IsString, OptionalArrayOfType, OptionalSimpleKey} from "@/app/components/common";


export function TestUserOk(){
    const a: UserData = {
        _id: "(EMAIL 1)",
        perms: {projects:["projA","projB"]},
    }
    return a
}

export function TestUserOk2(){
    const a: UserData = {
        _id: "(EMAIL 2)",
        perms: {projects:["projA","projB"]},
    }
    return a
}

export interface UserData {
    _id: string // This is username
    perms?: UserPerms
}

export interface UserPerms {
    admin?: boolean
    projects?: string[]
}

export function IsValidUserPerms(input: any): boolean {
    return (
        typeof input === 'object' &&
        OptionalSimpleKey('admin', input, "boolean") &&
        OptionalArrayOfType('projects', input, IsString)
    )
}

// Likely will never need a closeable single selector for this