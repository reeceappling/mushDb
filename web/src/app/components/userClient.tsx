'use client'

import React, {Dispatch, SetStateAction, useContext, useEffect, useState} from "react";
import ID from "@/app/components/formSubcomponents/id";
import {
    CheckArrayType, clientPostRequestHeaders,
    DisplayFormWrapper,
    DisplayInput, DoUpdateRequest, ErrHandler, FlexedArea, FlexedSinglesGroup,
    HandleJsonResponse,
    OptionalKey, updateApiUrlFor, validatorForAssertion,
} from "@/app/components/common";
import {ErrorDisplay} from "@/app/components/formSubcomponents/commonClient";
import {BaseExternalUrl} from "@/app/components/Constants";
import {IsValidUserPerms, UserData, UserPerms} from "@/app/components/userServer";
import {SelectorResetsOnSelectForCustom} from "@/app/components/selector";
import {MarshalAcl} from "@/app/components/accessControlClient";
import {AssertTransfer} from "@/app/components/transferClient";
import {allCookies, CookiesContext} from "@/app/components/formSubcomponents/cookiesContext/cookies";

export function AssertUser(input: any): asserts input is UserData {
    if (typeof input !== 'object') {
        throw new Error('Input is not an object! Input is ' + typeof input);
    }
    // required simple keys
    let requiredSimpleKeys = new Map<string, string>([
        ['_id', 'string'],
    ])
    for (let [key, expType] of requiredSimpleKeys) {
        if (!(key in input && typeof input[key] === expType)) {
            throw new Error('Transfer assertion failure: ' + key + 'was not type ' + expType + '. Was ' + (typeof input[key]));
        }
    }
    // // optional simple keys
    // let optionalSimpleKeys = new Map<string, string>([
    //     ['googleId', 'string'],
    // ])
    // for (let [key, expType] of optionalSimpleKeys) {
    //     if (!OptionalSimpleKey(key, input, expType)) {
    //         throw new Error('Transfer assertion failure: optional key ' + key + ' was not valid');
    //     }
    // }
    // complex optional keys
    let complexOptionalKeys = new Map<string, (v: any) => boolean>([
        ['perms', IsValidUserPerms]
    ])
    for (let [key, validator] of complexOptionalKeys) {
        if (!OptionalKey(key, input, validator)) {
            throw new Error('Plate assertion failure: optional key ' + key + ' was not valid');
        }
    }
    // // complex optional array keys
    // let complexOptionalArrayKeys = new Map<string, (v: any) => boolean>([
    //     ['notes', IsValidNote],
    // ])
    // for (let [key, validator] of complexOptionalArrayKeys) {
    //     if (!OptionalArrayOfType(key, input, validator)) {
    //         throw new Error('Transfer assertion failure: optional array key ' + key + ' was not valid');
    //     }
    // }
    return
}

// TODO: GetUserIdForEmail?

export default function UserDisplay(
    {
        id, readonly, data, headerLevel, isTopLevel
    }: DisplayInput) {
    try {

        AssertUser(data)
        const [initial, setInitial] = useState(data)
        const [err, setErr] = useState<string | undefined>()
        const [perms, setPerms] = useState<UserPerms>(initial.perms || {admin: true, projects: []}) // TODO: SETPERMS
        const updateInitial = (updated: UserData) => {
            setInitial(updated)
            setPerms(updated.perms || {admin: true, projects: []})
        }
        const cookies = useContext(CookiesContext)
        const userSubmit = () => {
            if ((!perms.admin && (initial.perms === undefined || initial.perms.admin)) || (perms.admin && (initial.perms && initial.perms.admin === false))) { // TODO: ensure ok
                const body: any = perms
                DoUpdateRequest("user",initial._id, body, AssertUser, allCookies(cookies))
                    .then(v=>{
                        updateInitial(new UserData(v))
                    })
                    .catch(e=>{
                        setErr(JSON.stringify(e))
                    })
            }
        }
        return (

            <DisplayFormWrapper entryType={"user"}>
                <ErrorDisplay err={err} headerLevel={headerLevel}/>
                <ID id={data._id} txt={"User"} entryType={"user"}/>
                <FlexedArea>
                    <FlexedSinglesGroup>{/*TODO: ALL THESE GROUPS!*/}
                    </FlexedSinglesGroup>
                </FlexedArea>
                {/* TODO: USERNAME, EMAIL, GOOGLE ID*/}
                {/* TODO: DISPLAY PERMS! (admin, projects)*/}
                {/*<UserPermsArea admin={perms.admin} projectNames={perms.projectNames} setAdmin={(adm?:boolean)=>{*/}
                {/*    let newPerms = {...perms}*/}
                {/*    newPerms.admin=adm*/}
                {/*    setPerms(newPerms)*/}
                {/*}}/>*/}
                {readonly ? null : <div><button className={"bottomButton greenButton"} onClick={(e)=>{
                    e.stopPropagation();
                    userSubmit()
                }}>{"Update"}</button></div>}
                {/* TODO: unlikely to need: <OnViewCreatorsQuadColArea OnViewCreators={ovcs} readonly={readonly}/> TODO: where to put?*/}
            </DisplayFormWrapper>
        )
    } catch (err) {
        return <div>{"ERROR: Transfer data format incorrect: " + err}</div>
    }
}

// TODO: MOVE!
export function HandleErr(error: any, setErr: Dispatch<SetStateAction<string | undefined>>) {
    if (error instanceof Error) {
        console.log(error.message) // TODO: del
        setErr(error.message)
    } else {
        console.log(JSON.stringify(error)) // TODO: del
        setErr(JSON.stringify(error))
    }
}

// TODO: move? keep on client because user can have different accesses
function getAllOptions<T>(itemType: string, assertEntry: (input: any) => void) {
    return fetch(BaseExternalUrl + "/db/list/" + itemType, {
        method: "GET",
        headers: clientPostRequestHeaders,
    }).then(HandleJsonResponse)
        .then((data) => {
            console.log("handling user selector response") // TODO: del
            if (!Array.isArray(data)) {
                console.log("db users response was not an array") // TODO: del
                throw "db users response was not an array"
            }
            if (!CheckArrayType(data, validatorForAssertion(assertEntry))) {
                throw "Error validating db users response"
            }
            return data as T[]
        })
}

// TODO: UserListPageTable


// TODO: ensure userSelector works as intended!
export function UserSelector(inp: {
    onSelect: (User: UserData) => void
    blacklist?: string[]
}) {
    // TODO: does this selector need depth?
    const [loading, setLoading] = useState(true)
    const [users, setUsers] = useState<UserData[]>([])
    const [err, setErr] = useState<string | undefined>(undefined)
    useEffect(() => { // TODO: FIX THIS SO IT GOES TO GET THE ACTUAL USERS THAT EXIST!!!! getUserIdsList endpoint!
        getAllOptions<UserData>("users", AssertUser).then((usrs) => { // TODO: REENABLE!!!
            setUsers(usrs)
            setErr(undefined)
            setLoading(false)
        }).catch((error) => {
            HandleErr(error, setErr) // TODO: use this everywhere!
        });
        // const usrs:UserData[] = [
        //     {_id: "userWithPerms", perms: {admin: false, projects: ["a","b","c"]}},
        //     {_id: "emptyPerms", perms: {}},
        //     {_id: "noPerms"},
        // ]
        // setUsers(usrs)
        // setLoading(false)
        // setErr(undefined)
    }, []);
    if (loading) {
        return <div>{"Loading users selector"}</div>
    }
    const opts = () => {
        return users.filter(pToFilter => {
            return (inp.blacklist || []).indexOf(pToFilter._id) == -1 && pToFilter._id !== ""
        })
    }
    return <div>
        <ErrorDisplay err={err}/>
        <SelectorResetsOnSelectForCustom options={opts()} updateParent={ud => {
            console.log("selected: ", ud._id) // TODO: fix
            inp.onSelect(ud)
        }} stringFor={(ud) => {
            return ud._id
        }}/>
    </div>
}