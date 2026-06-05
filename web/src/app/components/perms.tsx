import {BaseExternalUrl} from "@/app/components/Constants";
import {CheckArrayType, clientPostRequestHeaders, HandleTxtResponse, IsString} from "@/app/components/common";

interface CookieValues {
    SessionId?: string;
}

export interface ProjectWithPerm {
    projectName: string;
    canWrite?: boolean;
}



export interface SimplePerm<T> {
    entry: T
    canWrite?: boolean
}

export type UserWithPerm = SimplePerm<UserIdPair>

export interface EntryPerms {
    userPerms: UserPermSubset
    projectPerms: ProjectPermSubset
    blanketPerms: number
}

export class EntryPermsObject {
    public userPerms: UserPermSubset
    public projectPerms: ProjectPermSubset
    public blanketPerms: number
    constructor(userPerms: UserPermSubset, projectPerms: ProjectPermSubset, blanketPerms: number ) {
        this.userPerms = userPerms
        this.projectPerms = projectPerms
        this.blanketPerms = blanketPerms
    }
    doAThing() { // TODO: rename

    }
}

export function AssertEntryPerms(input: any): EntryPermsObject  { // TODO: USE
    if(IsValidEntryPerms(input)){
        const eps = input as EntryPerms
        return new EntryPermsObject(eps.userPerms, eps.projectPerms, eps.blanketPerms)
    }
    throw new Error("not a valid entry perms object")
}

export type ProjectPerms = { // TODO: replace with a Map<string, boolean> like in the go!
    users:   UserPermSubset
    blanket: number // 0 is none, 1 is read, 2 is write
}

export type ProjectPermSubset = {
    ids:      string[] // projectNames
    canWrite: boolean[]
}

export function filterProjectPermSubset(ppss: ProjectPermSubset, removeReads: boolean){
    const out: ProjectPermSubset = {...ppss}
    if(!removeReads){
        return out
    }
    let toKeep: number[] = []

    for(let i=0;i<ppss.ids.length;i++){
        if(ppss.canWrite){
            toKeep = [...toKeep, i]
        }
    }
    out.ids = toKeep.map((idx,_)=>{
        return ppss.ids[idx]
    })
    out.canWrite = toKeep.map((idx,_)=>{
        return ppss.canWrite[idx]
    })
    return out
}

export type UserPermSubset = {
    ids:      UserIdPair[]
    canWrite: boolean[]
}

export type UserIdPair = {
    id: string // AltCollId (binary)
    val: string // Email or username
}

export function IsValidUserIdPair(input: any)  {
    return (
        typeof input === 'object' &&
        'id' in input && typeof input.id === 'string' &&
        'val' in input && typeof input.val === 'string'
    )
}

export function IsValidProjectPerms(input: any)  {
    return (
        typeof input === 'object' &&
        'blanket' in input && typeof input.blanket === 'number' &&
        'users' in input && IsValidEntryPermSubset(input.users,IsValidUserIdPair)
    )
}

export function defaultEntryPerms(): EntryPerms {
    return {userPerms:{ids:[],canWrite:[]},projectPerms:{ids:[],canWrite:[]},blanketPerms:1}
}

export function IsValidEntryPerms(input: any): boolean {
    return (
        typeof input === 'object' &&
        'blanketPerms' in input && typeof input.blanketPerms === 'number' &&
        'userPerms' in input && typeof IsValidEntryPermSubset(input.userPerms, IsValidUserIdPair)&&
        'projectPerms' in input && IsValidEntryPermSubset(input.projectPerms, IsString)
    )
}

export function IsValidEntryPermSubset(input: any, idValidator: (v: any)=>boolean): boolean {
    return (
        typeof input === 'object' &&
        'ids' in input && Array.isArray(input.ids) && CheckArrayType(input.ids,idValidator) &&
        'canWrite' in input && Array.isArray(input.canWrite) && CheckArrayType(input.canWrite,(a: any)=>{return typeof a === 'boolean'}) &&
        input.ids.length === input.canWrite.length
    )
}


interface permSelectorInp {
    dontShowBelow?: boolean
    canWrite?: boolean,
    onChange: (v?: boolean)=>void
}

export function PermissionSelector({dontShowBelow,canWrite,onChange}:permSelectorInp){
    const onSelect = (val: string)=>{
        switch(val){
            case("write"):
                onChange(true)
                break;
            case("read"):
                onChange(false)
                break;
            default:
                onChange(undefined)
        }
    }
    const valueFor = (opt: boolean | undefined) =>{
        switch(opt){
            case undefined:
                return "none"
            case true:
                return "write"
            case false:
                return "read"
            default:
                return "none"
        }
    }
    const optionsFor = (dontShowBelow: boolean | undefined) =>{
        switch(dontShowBelow){
            case undefined:
                return ["write","read","none"]
            case false:
                return ["write","read"]
            case true:
                return ["write"]
        }
    }

    return (
        <select className={"tailwindSelector"} value={valueFor(canWrite)} onChange={(e)=>onSelect(e.currentTarget.value)}>
            {optionsFor(dontShowBelow).map((str: string)=>{
                return <option key={str/* TODO: ensure ok*/} value={str}>{str}</option>
            })}
        </select>
    )
}

export async function GetUserByString(/*sessionId?: string, */nameOrEmail?: string):Promise<string>{
    if(!nameOrEmail){
        return new Promise<string>((_,reject)=>{
            reject("no current name or email")
        })
    }
    const encodedNameOrEmail = encodeURIComponent(nameOrEmail)
    // if(!sessionId){
    // throw "missing session"
    // }
    return await fetch(BaseExternalUrl+"/idForUserOrEmail/"+encodedNameOrEmail, { // TODO: ensure works like we want!
        method: "GET",
        headers: clientPostRequestHeaders,
        credentials: 'include' // To include cookies
    }).then(HandleTxtResponse)
}

export function UserPermsArea(
    {
        admin,
        setAdmin,
        projectNames
    }:{
        admin?: boolean,
        setAdmin: (b?: boolean)=>void
        projectNames: string[],
    }
) {
    return (
        <div>
            <div>
                {"Is Admin?"}
                <input type="checkbox" checked={admin} onClick={()=>{
                   if(admin){
                       setAdmin(undefined)
                   } else {
                       setAdmin(true)
                   }
                }}/>
            </div>
            <div>
                {"Projects: "+ projectNames.join(",")/* TODO: be able to drop projects?*/}
            </div>
        </div>
    )
}