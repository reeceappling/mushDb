import {Note} from "@/app/components/formSubcomponents/notes";
import {EntryPerms, IsValidEntryPermSubset, IsValidUserIdPair} from "@/app/components/perms";
import {IsBool, IsString, OptionalArrayOfType, OptionalSimpleKey} from "@/app/components/common";


export function TestUserOk(){
    return new UserData({
        _id: "(EMAIL 1)",
        perms: {projects:["projA","projB"]},
    })
}

export function TestUserOk2(){
    return new UserData({
        _id: "(EMAIL 2)",
        perms: {projects:["projA","projB"]},
    })
}

export interface UserData {
    _id: string // This is username
    perms?: UserPerms
}
export class UserData {
    // Accept a single object containing the fields
    constructor(init?: Partial<UserData>) {
        // Dynamically map the object fields onto the class instance
        Object.assign(this, init);
    }

    public getId(): string {
        return this._id
    }
    public entryType(): string {
        return "user"
    }
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